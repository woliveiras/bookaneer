---
name: go-service
description: "Scaffold a new Go service package following Bookaneer's architecture. Use when creating a new domain service, adding a new integration (metadata provider, download client, indexer, notification channel), or adding a new API handler. Generates code following the project's layered architecture, interface patterns, and conventions."
---

# Go Service Scaffolding

## When to Use
- Creating a new domain service (e.g., `internal/core/series/`)
- Adding a new integration (e.g., `internal/metadata/hardcover/`)
- Adding a new API handler group (e.g., `api/v1/handler/series.go`)
- Adding a new download client or notification channel

## Procedure

### 1. Determine the Type

**Domain Service** (internal/core/):
- Has business logic, owns SQL queries
- Pattern: model → service → handler

**Integration** (internal/metadata/, search/, download/, notification/):
- Implements an interface defined by its consumer
- Pattern: interface (consumer side) → implementation → registration

**Handler** (api/v1/handler/):
- Thin HTTP layer, no business logic
- Pattern: parse request → call service → format response

### 2. Scaffold Domain Service

Create this file structure:
```
internal/core/{name}/
├── model.go     # Domain types
├── service.go   # Business logic + SQL
├── errors.go    # Sentinel errors
└── service_test.go
```

**model.go**:
```go
package {name}

type {Name} struct {
    ID        int64  `json:"id"`
    // fields...
    CreatedAt string `json:"createdAt"`
    UpdatedAt string `json:"updatedAt"`
}
```

**errors.go**:
```go
package {name}

import "errors"

var (
    ErrNotFound  = errors.New("{name} not found")
    ErrDuplicate = errors.New("{name} already exists")
)
```

**service.go**:
```go
package {name}

import (
    "context"
    "database/sql"
    "fmt"
)

type Service struct {
    db *sql.DB
}

func New(db *sql.DB) *Service {
    return &Service{db: db}
}

func (s *Service) FindByID(ctx context.Context, id int64) (*{Name}, error) {
    var item {Name}
    err := s.db.QueryRowContext(ctx,
        "SELECT id, ... FROM {table} WHERE id = ?", id,
    ).Scan(&item.ID, ...)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("find {name} %d: %w", id, err)
    }
    return &item, nil
}
```

### 3. Scaffold Integration

```
internal/{category}/{provider}/
├── provider.go      # Implementation
└── provider_test.go
```

**Interface** (defined in consumer package):
```go
// internal/{category}/provider.go
package {category}

type Provider interface {
    Method(ctx context.Context, params...) (Result, error)
}
```

**Implementation**:
```go
package {provider}

import "bookaneer/internal/{category}"

var _ {category}.Provider = (*{Provider})(nil)

type {Provider} struct {
    client  *http.Client
    baseURL string
}

func New(client *http.Client) *{Provider} {
    return &{Provider}{
        client:  client,
        baseURL: "https://...",
    }
}
```

### 4. Scaffold Handler

```go
// api/v1/handler/{name}.go
package handler

type {Name}Handler struct {
    svc *{name}.Service
}

func New{Name}Handler(svc *{name}.Service) *{Name}Handler {
    return &{Name}Handler{svc: svc}
}

func (h *{Name}Handler) Register(g *echo.Group) {
    g.GET("/{name}", h.List)
    g.GET("/{name}/:id", h.GetByID)
    g.POST("/{name}", h.Create)
    g.PUT("/{name}/:id", h.Update)
    g.DELETE("/{name}/:id", h.Delete)
}

func (h *{Name}Handler) GetByID(c echo.Context) error {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
    }
    item, err := h.svc.FindByID(c.Request().Context(), id)
    if err != nil {
        if errors.Is(err, {name}.ErrNotFound) {
            return echo.NewHTTPError(http.StatusNotFound)
        }
        return fmt.Errorf("get {name}: %w", err)
    }
    return c.JSON(http.StatusOK, item)
}
```

### 5. Wire in main.go

```go
// cmd/bookaneer/main.go — inside run()
{name}Svc := {name}.New(db)
{name}Handler := handler.New{Name}Handler({name}Svc)
{name}Handler.Register(v1)
```

### 6. Checklist

- [ ] Model types defined with JSON tags
- [ ] Sentinel errors defined
- [ ] Service accepts `*sql.DB` and dependencies via constructor
- [ ] All methods accept `context.Context` as first param
- [ ] SQL queries use parameterized `?` placeholders
- [ ] Handler validates input before calling service
- [ ] Handler maps domain errors to HTTP status codes
- [ ] Interface compliance verified: `var _ Interface = (*Impl)(nil)`
- [ ] Test file created with table-driven tests
- [ ] Wired in main.go
