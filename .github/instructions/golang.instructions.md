---
description: "Use when writing, reviewing, or refactoring Go code. Covers Go code quality, idiomatic patterns, error handling, naming, testing, linting, and performance. Based on Effective Go and Uber Go Style Guide."
applyTo: "**/*.go"
---

# Go Code Quality

Standards based on [Effective Go](https://go.dev/doc/effective_go) and [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

## Formatting & Imports

- `gofmt` is non-negotiable — all code must be formatted
- Use `goimports` to manage imports automatically
- Two import groups: stdlib, then everything else (separated by blank line)
- No import aliases unless there's a name conflict

## Naming

- `MixedCaps` / `mixedCaps` — never underscores
- Package names: lowercase, single word, no plural (`book` not `books`, `auth` not `authentication`)
- No `Get` prefix on getters: `book.Title()` not `book.GetTitle()`
- Interface names: one-method interfaces end in `-er` (`Reader`, `Scanner`, `Downloader`)
- Error variables: `ErrNotFound` (exported), `errNotFound` (unexported)
- Error types: `NotFoundError` (suffix `Error`)
- Unexported package-level vars: prefix with `_` (`_defaultTimeout`)

## Error Handling

- Handle every error: `if err != nil` is never optional
- Wrap errors with context: `fmt.Errorf("find book %d: %w", id, err)`
- Handle errors once: either log OR return, never both
- Use `%w` for errors callers need to match, `%v` to obfuscate
- Error messages: lowercase, no "failed to" prefix, concise
- Use sentinel errors (`var ErrNotFound = errors.New(...)`) for errors callers match with `errors.Is`
- Use custom types (`type NotFoundError struct{...}`) for errors callers match with `errors.As`
- Never panic in library code: return errors. Panic only for truly unrecoverable states during init

## Interfaces

- Accept interfaces, return structs
- Define interfaces where they're used (consumer side), not where they're implemented
- Small interfaces: 1-2 methods are ideal
- Verify compliance at compile time:
  ```go
  var _ http.Handler = (*MyHandler)(nil)
  ```

## Concurrency

- Share by communicating, don't communicate by sharing
- Every goroutine must have a clear shutdown mechanism (`context.Context`, `done` channel)
- Never fire-and-forget goroutines: always track lifetime
- Use `sync.WaitGroup` for multiple goroutines, `chan struct{}` for single
- Protect shared state with `sync.Mutex` (not embedding) or channels
- Always run `go test -race ./...` to detect races
- Channel size: 0 (synchronous) or 1. Larger buffers need justification

## Structs & Types

- Use field names when initializing structs: `User{Name: "x"}` not `User{"x"}`
- Omit zero-value fields in struct literals
- Use `var x MyStruct` for zero-value structs, `&MyStruct{...}` for non-zero
- Embedding: only when it provides real functional benefit
- Never embed mutex in public structs — use a named field `mu sync.Mutex`
- Use field tags for all marshaled structs: `json:"name"`

## Functions

- `run()` pattern in main: `func main() { if err := run(); err != nil { log.Fatal(err) } }`
- `os.Exit` / `log.Fatal` only in `main()`
- Group functions by receiver, sorted in call order
- Exported functions first, after type definition
- Reduce nesting: handle errors first, return early
- No unnecessary else: `x := 10; if b { x = 100 }`

## Performance (hot paths only)

- Prefer `strconv` over `fmt` for primitive conversions
- Avoid repeated `[]byte(string)` conversions
- Pre-allocate slices and maps with `make(T, 0, capacity)` when size is known
- Use `strings.Builder` for string concatenation in loops

## Testing

- Table-driven tests with `testify` assertions
- Use `give`/`want` prefixes for test case fields
- `t.Parallel()` where safe
- `t.Fatal` / `t.FailNow` for setup failures, never `panic`
- Use subtests: `t.Run("case name", func(t *testing.T) {...})`
- Don't mock what you don't own — wrap external dependencies behind interfaces
- Test public API surface, not internal implementation
- Fuzzing for parser/input handling functions

## Linting

Run `golangci-lint` with at minimum:
- `errcheck` — ensure errors are handled
- `govet` — suspicious constructs
- `staticcheck` — static analysis
- `revive` — style mistakes
- `goimports` — import formatting
- `gosec` — security issues

## Anti-patterns to Avoid

- `init()` functions (except for driver registration)
- Mutable global variables — use dependency injection
- `interface{}` / `any` without type assertions
- Named returns except for short functions or `defer` error handling
- Embedding types in public structs without justification
- Using built-in names as identifiers (`error`, `string`, `len`)
- Fire-and-forget goroutines
- `panic` in library code
