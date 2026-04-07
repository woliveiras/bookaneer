---
description: "Use when writing or reviewing code for security concerns, handling user input, authentication, authorization, cryptography, API security, secrets management, or OWASP compliance. Applies to all backend and frontend code."
applyTo: "**/*.go, **/*.ts, **/*.tsx, Dockerfile, docker-compose*.yml"
---

# Application Security — Security by Design

This project follows security-by-design principles aligned with the OWASP Top 10:2025.

## A01: Broken Access Control

- Every API endpoint requires authentication (API key header or session) except `/auth/login`
- Enforce authorization checks at the service layer, not just the handler
- Use allowlists for file path operations — never allow user input to traverse directories
- Deny by default: if a permission check is missing, the request is denied
- Validate that users can only access their own resources (e.g., reading progress)

## A02: Security Misconfiguration

- Never expose stack traces or internal error details in API responses
- Return generic error messages to clients; log detailed errors server-side
- Set security headers on all responses:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `Content-Security-Policy` (strict, no inline scripts)
  - `Strict-Transport-Security` when behind HTTPS
- Disable directory listing on static file serving
- Default config is secure: auth enabled, random API key generated on first boot

## A03: Software Supply Chain Failures

- Pin Go dependencies with `go.sum` verification
- Pin frontend dependencies with `pnpm-lock.yaml`
- Run `govulncheck ./...` in CI to scan for known vulnerabilities
- Run `npm audit` or equivalent for frontend dependencies
- Verify Docker base images with digests in production builds
- Minimize dependencies — prefer stdlib over third-party when reasonable

## A04: Cryptographic Failures

- Passwords: bcrypt with cost ≥ 12, never store plaintext
- API keys: generated with `crypto/rand`, minimum 32 bytes, hex-encoded
- Use `crypto/rand` for all random values — never `math/rand` for security
- Store secrets (download client passwords, SMTP passwords) encrypted at rest or ensure the data directory has restricted permissions
- Use HTTPS for all external API calls (metadata providers, indexers, download clients)

## A05: Injection

- **SQL**: Always use parameterized queries (`?` placeholders). Never concatenate user input into SQL strings
  ```go
  // CORRECT
  db.Query("SELECT * FROM books WHERE id = ?", id)
  
  // NEVER DO THIS
  db.Query("SELECT * FROM books WHERE id = " + id)
  ```
- **Command injection**: Never pass user input to `os/exec` without validation
- **Path traversal**: Validate and sanitize all file paths. Use `filepath.Clean()` and verify the result is within allowed directories
- **XSS**: React escapes by default. Never use `dangerouslySetInnerHTML`. Sanitize any HTML from metadata providers before rendering
- **OPDS/XML**: Use `encoding/xml` marshaling — never build XML via string concatenation

## A06: Insecure Design

- Threat model: the application runs on a home network but may be exposed via reverse proxy
- Rate-limit login attempts to prevent brute force
- Validate all input at API boundaries: type, length, range, format
- File uploads (book imports): validate file type by magic bytes, not just extension
- Limit request body size to prevent memory exhaustion

## A07: Authentication Failures

- Passwords: minimum 8 characters, bcrypt hashed
- API keys: 64-character hex strings from `crypto/rand`
- Session tokens: HTTP-only, Secure, SameSite=Strict cookies (when using session auth)
- Lock accounts after repeated failed login attempts (configurable threshold)
- Force API key regeneration if compromised

## A08: Software or Data Integrity Failures

- Verify checksums of downloaded ebooks when available
- Use Go's `embed.FS` for frontend assets — immutable at build time
- Sign release binaries (future: GitHub Actions artifact signing)
- Validate webhook payloads with HMAC when receiving external callbacks

## A09: Security Logging and Alerting Failures

- Log all authentication events (login, logout, failed attempts) with timestamp and IP
- Log authorization failures (access denied)
- Never log secrets, passwords, API keys, or tokens
- Structured logging (JSON format in production) for easy parsing
- Include request ID in all log entries for correlation

## A10: Mishandling of Exceptional Conditions

- Never panic in request handlers — always return errors
- Use `recover()` middleware in Echo to catch unexpected panics and return 500
- Handle all error returns: `if err != nil` is mandatory, not optional
- Timeout all external HTTP calls (metadata, indexer, download client) with `context.WithTimeout`
- Graceful shutdown: drain in-flight requests, stop scheduler, close DB

## Go-Specific Security Patterns

```go
// Always validate path operations
func safePath(root, userPath string) (string, error) {
    cleaned := filepath.Clean(userPath)
    full := filepath.Join(root, cleaned)
    if !strings.HasPrefix(full, filepath.Clean(root)+string(os.PathSeparator)) {
        return "", fmt.Errorf("path traversal attempt: %s", userPath)
    }
    return full, nil
}

// Always use parameterized queries
func (r *BookRepo) FindByID(ctx context.Context, id int64) (*Book, error) {
    row := r.db.QueryRowContext(ctx, "SELECT id, title FROM books WHERE id = ?", id)
    // ...
}

// Always use crypto/rand for secrets
func generateAPIKey() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("generate api key: %w", err)
    }
    return hex.EncodeToString(b), nil
}
```

## Frontend Security

- No inline scripts or styles — strict CSP
- Sanitize metadata from external sources before rendering
- Use `httpOnly` cookies for session management
- Never store tokens in `localStorage` — use memory or secure cookies
- Validate and escape user inputs in forms before submission
