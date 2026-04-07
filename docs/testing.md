# Testing Guide

This document describes how to test Bookaneer at different levels.

## Quick Start

```bash
# Run all tests
make test

# Run with race detector (recommended for CI)
make test-race

# Run with coverage report
make test-cover
```

---

## Test Structure

Tests are organized alongside the code they test:

```
internal/
├── auth/
│   ├── service.go
│   └── service_test.go      # Unit tests for auth service
├── database/
│   ├── db.go
│   └── db_test.go           # Integration tests with SQLite
api/v1/
├── handler/
│   ├── system.go
│   └── system_test.go       # HTTP handler tests
```

---

## Unit Tests

Unit tests focus on business logic in isolation.

### Writing Unit Tests

```go
package auth_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/woliveiras/bookaneer/internal/auth"
    "github.com/woliveiras/bookaneer/internal/database"
)

func TestService_Authenticate(t *testing.T) {
    t.Parallel()

    // Setup in-memory database
    db, err := database.Open(":memory:")
    require.NoError(t, err)
    defer db.Close()

    // Run migrations
    // ...

    svc := auth.New(db)

    tests := []struct {
        name     string
        username string
        password string
        wantErr  error
    }{
        {
            name:     "valid credentials",
            username: "admin",
            password: "secret",
            wantErr:  nil,
        },
        {
            name:     "invalid password",
            username: "admin",
            password: "wrong",
            wantErr:  auth.ErrInvalidCredentials,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            
            user, err := svc.Authenticate(context.Background(), tt.username, tt.password)
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, user)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, user)
            }
        })
    }
}
```

### Best Practices

- Use `t.Parallel()` for independent tests
- Use table-driven tests with `give`/`want` naming
- Use `require` for setup failures, `assert` for test assertions
- Test public API surface, not internal implementation
- Use `:memory:` SQLite for fast, isolated database tests

---

## Integration Tests

Integration tests verify components work together.

### API Handler Tests

```go
package handler_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
    "github.com/woliveiras/bookaneer/api/v1/handler"
    "github.com/woliveiras/bookaneer/internal/config"
)

func TestSystemHandler_Health(t *testing.T) {
    e := echo.New()
    cfg := config.DefaultConfig()
    h := handler.NewSystemHandler("test", "now", cfg)

    req := httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    err := h.Health(c)

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Body.String(), `"status":"ok"`)
}
```

### Database Integration Tests

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    
    db, err := database.Open(":memory:")
    require.NoError(t, err)
    
    // Run migrations
    err = database.Migrate(db, migrationsFS, "migrations")
    require.NoError(t, err)
    
    t.Cleanup(func() { db.Close() })
    
    return db
}
```

---

## End-to-End Tests

E2E tests verify the full system from HTTP to database.

```go
func TestE2E_AuthFlow(t *testing.T) {
    // Start test server
    cfg := config.DefaultConfig()
    cfg.DataDir = t.TempDir()
    
    // ... setup server ...
    
    // Test: Create user → Login → Access protected endpoint
    // ...
}
```

---

## Frontend Tests

```bash
cd web

# Run tests
pnpm test

# Run with coverage
pnpm test:coverage

# Run in watch mode
pnpm test:watch
```

### Component Tests (Vitest)

```tsx
// src/components/Button.test.tsx
import { render, screen } from '@testing-library/react'
import { Button } from './Button'

describe('Button', () => {
  it('renders children', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByRole('button')).toHaveTextContent('Click me')
  })
})
```

---

## CI Integration

The GitHub Actions workflow runs:

1. `make lint` — golangci-lint + eslint
2. `make test-race` — Go tests with race detector
3. `pnpm --prefix web test` — Frontend tests
4. `make docker` — Docker build (ensures everything compiles)

---

## Test Coverage

Generate coverage reports:

```bash
make test-cover
open data/coverage.html
```

Current coverage targets:
- `internal/core/*`: > 80%
- `internal/auth`: > 90%
- `api/v1/handler/*`: > 70%

---

## Debugging Tests

```bash
# Run specific test with verbose output
go test -v -run TestService_Authenticate ./internal/auth/

# Run with race detector on specific package
go test -race ./internal/auth/

# Show test coverage for specific package
go test -cover ./internal/auth/
```
