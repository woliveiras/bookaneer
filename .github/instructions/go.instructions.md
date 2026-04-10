---
applyTo: "**/*.go"
---

# Go Conventions

## Error Handling
- Always wrap errors with context: `fmt.Errorf("action %d: %w", id, err)`
- Define sentinel errors per package: `var ErrNotFound = errors.New("not found")`
- Return errors explicitly — do not panic in library code
- Check every error — no `_` for error returns

## Interfaces
- Keep interfaces small (1-3 methods)
- Define interfaces where they are consumed, not where they are implemented
- Verify implementation at compile time: `var _ Interface = (*Impl)(nil)`

## Functions
- Accept `context.Context` as the first parameter
- Use dependency injection via constructors: `func New(db *sql.DB) *Service`
- Prefer returning values over mutating parameters

## Testing
- Table-driven tests with `t.Run()` subtests
- Use `testify/assert` and `testify/require` for assertions
- Use `t.TempDir()` for filesystem isolation
- Create test helpers in a `testutil` package
