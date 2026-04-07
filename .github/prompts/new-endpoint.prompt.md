---
description: "Scaffold a new REST API endpoint: handler, route registration, and test file. Use when adding a new endpoint to the API."
agent: "agent"
argument-hint: "Endpoint name and HTTP method (e.g., 'GET /api/v1/series')"
---

Scaffold a new API endpoint for Bookaneer following the project conventions.

## Requirements

1. **Handler** in `api/v1/handler/` — thin HTTP layer, no business logic:
   - Parse and validate request (path params, query params, body)
   - Call the appropriate service method
   - Return JSON response with proper status code
   - Follow [architecture instructions](../instructions/architecture.instructions.md)

2. **Route registration** in `api/v1/router.go`:
   - Add the route under the correct group
   - Apply auth middleware

3. **Handler test** in `api/v1/handler/` — table-driven tests with `testify`:
   - Happy path
   - Validation error (400)
   - Not found (404)
   - Follow [Go instructions](../instructions/golang.instructions.md)

## Conventions

- Use Echo's `c.Bind()` for request parsing
- Return `echo.NewHTTPError(status, message)` for errors
- Use `context.Context` from `c.Request().Context()`
- JSON field names are `camelCase`
- Pagination: `page` and `pageSize` query params, response includes `page`, `pageSize`, `totalItems`
