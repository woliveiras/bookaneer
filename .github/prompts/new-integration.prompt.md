---
description: "Scaffold a new external integration (metadata provider, indexer, download client, or notification channel). Use when adding support for a new external service."
agent: "agent"
argument-hint: "Integration type and provider (e.g., 'metadata provider for Hardcover')"
---

Scaffold a new external integration for Bookaneer.

## Steps

1. **Identify the category**: `metadata`, `search`, `download`, or `notification`
2. **Check if the interface exists** in `internal/{category}/provider.go` — create if it doesn't
3. **Create the implementation** following the interface + implementation pattern

## File Structure

```
internal/{category}/{provider}/
├── provider.go       # Implementation of the interface
└── provider_test.go  # Table-driven tests
```

## Rules

Follow [architecture instructions](../instructions/architecture.instructions.md):

- Interface is defined in the **consumer** package (e.g., `internal/metadata/provider.go`)
- Implementation lives in a sub-package (e.g., `internal/metadata/openlibrary/`)
- Verify interface compliance: `var _ {category}.Provider = (*{Provider})(nil)`
- Use `*http.Client` as dependency — never create your own
- Constructor: `func New(client *http.Client, opts ...Option) *{Provider}`
- All methods take `context.Context` as first parameter
- Wrap errors with context: `fmt.Errorf("{provider} search: %w", err)`
- Do NOT import `api/` packages — dependency flows inward
- Register the implementation in `cmd/bookaneer/main.go`
