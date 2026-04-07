# Phase 1 Architecture — Foundation

This document describes the architecture implemented in Phase 1 of Bookaneer.

## Overview

Phase 1 establishes the foundational infrastructure:

```
┌─────────────────────────────────────────────────────────────────┐
│                         Bookaneer                               │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────┐ │
│  │   React SPA     │    │   Echo HTTP     │    │   SQLite    │ │
│  │   (embedded)    │◄──►│    Server       │◄──►│   (WAL)     │ │
│  └─────────────────┘    └─────────────────┘    └─────────────┘ │
│                               │                                 │
│                         ┌─────┴─────┐                           │
│                         │   Auth    │                           │
│                         │  Service  │                           │
│                         └───────────┘                           │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Entry Point (`cmd/bookaneer/main.go`)

The main.go file is responsible for:

- Parsing command-line flags
- Loading configuration
- Opening the database connection
- Running migrations
- Wiring all dependencies together
- Starting the HTTP server
- Handling graceful shutdown

**Key patterns:**
- `run()` pattern for testability
- Embed migrations and frontend via `embed.FS`
- Signal handling for graceful shutdown

### 2. Configuration (`internal/config/`)

Configuration follows a layered approach:

```
Priority (highest to lowest):
1. Command-line flags
2. Environment variables (BOOKANEER_*)
3. config.yaml file
4. Default values
```

**Environment variables:**
- `BOOKANEER_PORT` — HTTP port (default: 9090)
- `BOOKANEER_DATA_DIR` — Data directory (default: ./data)
- `BOOKANEER_LOG_LEVEL` — Log level (debug/info/warn/error)
- `BOOKANEER_LIBRARY_DIR` — Library root folder

### 3. Database (`internal/database/`)

SQLite with WAL mode for concurrent reads:

```go
dsn := "file:path?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_synchronous=NORMAL"
db.SetMaxOpenConns(1)  // Single writer for SQLite
```

**Migrations:**
- SQL files in `migrations/` directory
- Embedded in binary via `embed.FS`
- Run automatically on startup via goose
- Schema defined in `001_initial_schema.sql`

### 4. Authentication (`internal/auth/`)

Two authentication methods:

1. **API Key** — For programmatic access
   - System API key stored in `config` table
   - User API keys stored in `users` table
   - Passed via `X-Api-Key` header or `?apikey=` query param

2. **Username/Password** — For UI login
   - Passwords hashed with bcrypt (cost=10)
   - Returns user's API key on successful login
   - Client stores API key for subsequent requests

**Auth flow:**
```
Client                          Server
  │                                │
  ├──POST /api/v1/auth/login──────►│
  │   {username, password}         │
  │                                │ Validate credentials
  │                                │ Return API key
  │◄──────────{apiKey}─────────────┤
  │                                │
  ├──GET /api/v1/books ───────────►│
  │   X-Api-Key: <key>             │
  │                                │ Validate key
  │◄──────────{data}───────────────┤
```

### 5. HTTP Layer (`api/v1/`)

Echo framework with middleware stack:

```go
e.Use(middleware.Recover())     // Panic recovery
e.Use(middleware.RequestID())   // Request tracing
e.Use(apimw.Logger())           // Structured logging
e.Use(middleware.CORS())        // CORS headers
```

**Endpoint groups:**
- `/api/v1/system/*` — Public (health, status)
- `/api/v1/auth/*` — Public (login)
- `/api/v1/*` — Protected (requires API key)

### 6. Frontend (`web/`)

React SPA with:
- **Vite** — Build tooling
- **TanStack Query** — Data fetching and caching
- **TanStack Router** — Type-safe routing (Phase 2)
- **Tailwind CSS v4** — Styling
- **shadcn/ui** — Component library
- **lucide-react** — Icons

**Embedding:**
```go
//go:embed all:web/dist
var webFS embed.FS
```

The frontend is embedded in the Go binary and served via Echo's static file handler with SPA fallback.

---

## File Structure

```
bookaneer/
├── cmd/bookaneer/
│   └── main.go              # Entrypoint, wiring
├── internal/
│   ├── auth/
│   │   └── service.go       # Auth business logic
│   ├── config/
│   │   └── config.go        # Configuration loading
│   └── database/
│       └── db.go            # SQLite connection, migrations
├── api/v1/
│   ├── handler/
│   │   ├── system.go        # Health, status endpoints
│   │   └── auth.go          # Login, logout endpoints
│   └── middleware/
│       └── middleware.go    # Logger, auth middleware
├── migrations/
│   └── 001_initial_schema.sql
├── web/
│   ├── src/
│   │   ├── App.tsx          # Main React component
│   │   ├── main.tsx         # React entry point
│   │   ├── index.css        # Tailwind styles
│   │   └── lib/
│   │       └── utils.ts     # shadcn utilities
│   ├── vite.config.ts
│   └── package.json
├── Dockerfile
├── Makefile
├── .air.toml
└── go.mod
```

---

## Data Flow

### Request lifecycle:

```
HTTP Request
    │
    ▼
┌───────────────┐
│   Middleware  │  Recover → RequestID → Logger → CORS → Auth
└───────┬───────┘
        │
        ▼
┌───────────────┐
│   Handler     │  Parse request, validate input
└───────┬───────┘
        │
        ▼
┌───────────────┐
│   Service     │  Business logic, SQL queries
└───────┬───────┘
        │
        ▼
┌───────────────┐
│   Database    │  SQLite (single connection)
└───────────────┘
```

### Dependency injection:

```go
// main.go — all dependencies wired at startup
db := database.Open(cfg.DatabasePath())
authSvc := auth.New(db)
authHandler := handler.NewAuthHandler(authSvc)
```

No global state. All dependencies passed explicitly via constructors.

---

## Security

### Implemented in Phase 1:

- ✅ API key authentication
- ✅ bcrypt password hashing (cost=10)
- ✅ Parameterized SQL queries (no injection)
- ✅ CORS headers
- ✅ Request ID for tracing

### Planned for Phase 2+:

- Session tokens with expiry
- CSRF protection
- Rate limiting
- Input validation middleware
- Audit logging

---

## Next Steps (Phase 2)

1. **Authors & Books CRUD** — Domain services
2. **Metadata providers** — OpenLibrary integration
3. **Library scan** — File system crawler
4. **Frontend routing** — TanStack Router setup
5. **WebSocket** — Real-time notifications
