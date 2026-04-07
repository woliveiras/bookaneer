---
description: "Use when making architectural decisions, creating new packages, defining interfaces, structuring services, or working with the internal/ package layout. Covers Bookaneer's layered architecture, dependency rules, and Go project structure patterns."
applyTo: "internal/**/*.go, api/**/*.go, cmd/**/*.go"
---

# Architecture & Go Project Structure

## Directory Layout Rules

```
cmd/bookaneer/main.go  → Only wiring: parse flags, init deps, start server
internal/              → All business logic lives here
  core/                → Domain models and services (pure logic, no HTTP)
  metadata/            → External metadata providers
  search/              → External indexer queries
  download/            → External download client integrations
  notification/        → External notification channels
  opds/                → OPDS catalog serving
  reader/              → Web reader API
  scheduler/           → Job queue and scheduler
  config/              → Configuration loading
  auth/                → Authentication and authorization
  database/            → DB connection, helpers, migration runner
api/v1/                → HTTP layer only (handlers, middleware, router, websocket)
web/                   → React frontend (embedded at build time)
migrations/            → SQL migration files
```

## Dependency Flow

Dependencies flow inward. The outer layer depends on the inner, never reversed.

```
HTTP handlers (api/) → Services (internal/core/) → Repository (internal/core/)
                     → Integrations (internal/metadata/, search/, download/)
                     → Database (internal/database/)
```

Rules:
- `internal/core/` NEVER imports `api/`
- `internal/core/` NEVER imports HTTP-related packages (`net/http`, `echo`)
- `api/v1/handler/` calls service methods, never writes SQL directly
- `internal/metadata/`, `search/`, `download/`, `notification/` are peers — they don't import each other
- `cmd/bookaneer/main.go` is the only place where all packages are wired together

## Service Pattern

Each domain package follows this structure:

```go
// internal/core/book/service.go
type Service struct {
    db     *sql.DB
    meta   metadata.Aggregator  // interface, not concrete type
    events chan<- Event          // for WebSocket notifications
}

func New(db *sql.DB, meta metadata.Aggregator, events chan<- Event) *Service {
    return &Service{db: db, meta: meta, events: events}
}

func (s *Service) FindByID(ctx context.Context, id int64) (*Book, error) {
    // SQL query directly in service — no ORM, no separate repository file
    // unless the queries grow large enough to warrant extraction
}
```

Key points:
- Constructor `New()` receives all dependencies explicitly — no global state
- Methods accept `context.Context` as first parameter
- Return domain types, not HTTP types
- SQL lives in the service (or extracted to a `queries.go` in the same package)

## Interface + Implementation Pattern

For external integrations (metadata, indexers, download clients, notifications):

```go
// internal/metadata/provider.go — interface defined where it's CONSUMED
type Provider interface {
    SearchAuthor(ctx context.Context, query string) ([]AuthorResult, error)
    SearchBook(ctx context.Context, query string) ([]BookResult, error)
    GetAuthor(ctx context.Context, foreignID string) (*Author, error)
    GetBook(ctx context.Context, foreignID string) (*Book, error)
}

// internal/metadata/openlibrary/provider.go — implementation
type OpenLibrary struct {
    client *http.Client
    baseURL string
}

var _ metadata.Provider = (*OpenLibrary)(nil) // compile-time check

func New(client *http.Client) *OpenLibrary {
    return &OpenLibrary{client: client, baseURL: "https://openlibrary.org"}
}
```

## Handler Pattern

HTTP handlers are thin: validate input → call service → format response.

```go
// api/v1/handler/book.go
type BookHandler struct {
    books *book.Service
}

func NewBookHandler(books *book.Service) *BookHandler {
    return &BookHandler{books: books}
}

func (h *BookHandler) GetByID(c echo.Context) error {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
    }

    b, err := h.books.FindByID(c.Request().Context(), id)
    if err != nil {
        if errors.Is(err, book.ErrNotFound) {
            return echo.NewHTTPError(http.StatusNotFound, "book not found")
        }
        return fmt.Errorf("get book %d: %w", id, err)
    }

    return c.JSON(http.StatusOK, b)
}
```

Rules:
- Parse and validate request parameters in the handler
- Pass `context.Context` from the request to service calls
- Map domain errors to HTTP status codes in the handler
- Return `echo.NewHTTPError` for client errors, `fmt.Errorf` for unexpected errors (Echo middleware converts to 500)
- No business logic in handlers

## Configuration

- Config loaded once at startup in `main.go`
- Passed as a struct to services that need it — not as a global
- Environment variables override config file values
- Sensitive defaults: auth enabled, random API key, bind to `0.0.0.0:9090`

## Background Jobs

- Jobs are queued by inserting rows into the `commands` table
- The scheduler goroutine polls every 1s for pending commands
- Each command runs in its own goroutine with a `context.Context`
- On startup: reset `running` commands to `queued` (crash recovery)
- Commands emit progress via WebSocket channel

## Error Domain

Define domain-specific sentinel errors in each package:

```go
// internal/core/book/errors.go
var (
    ErrNotFound     = errors.New("book not found")
    ErrDuplicate    = errors.New("book already exists")
    ErrInvalidISBN  = errors.New("invalid ISBN")
)
```

## Testing Architecture

- Unit tests: test services with a real SQLite in-memory DB (`file::memory:?cache=shared`)
- Integration tests: test handlers with `httptest` + real service + in-memory DB
- E2E tests: Playwright against a running instance
- External integrations: mock behind interfaces in unit tests, integration test with real APIs behind build tag `//go:build integration`
