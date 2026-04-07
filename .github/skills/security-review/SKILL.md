---
name: security-review
description: "Review code for security vulnerabilities aligned with OWASP Top 10:2025. Use when asked to audit, review, or check code for security issues including SQL injection, path traversal, authentication flaws, broken access control, and cryptographic failures in Go backend or React frontend code."
---

# Security Review

## When to Use
- Reviewing PRs or code changes for security issues
- Auditing existing code for vulnerabilities
- Checking that new endpoints follow security patterns
- Validating input handling and authentication flows

## Procedure

### 1. Identify Attack Surface
- List all user inputs: URL params, query strings, request body, headers, file uploads
- List all external data sources: metadata APIs, indexer responses, download client callbacks
- List all file system operations: library scan, book import, cover caching, backup

### 2. Check OWASP Top 10:2025

For each file/change, verify against these categories:

**A01 — Broken Access Control**
- [ ] All endpoints require authentication (API key or session)
- [ ] Authorization checked at service layer (user can only access own resources)
- [ ] No IDOR: user-supplied IDs validated against ownership
- [ ] File paths validated against root folder boundaries

**A02 — Security Misconfiguration**
- [ ] No stack traces or internal details leaked in API responses
- [ ] Security headers set (X-Content-Type-Options, X-Frame-Options, CSP)
- [ ] Default config is secure (auth enabled, random API key)

**A03 — Supply Chain**
- [ ] No new dependencies without justification
- [ ] Dependencies pinned in go.sum / pnpm-lock.yaml

**A04 — Cryptographic Failures**
- [ ] Passwords: bcrypt with cost ≥ 12
- [ ] Secrets: `crypto/rand` not `math/rand`
- [ ] No plaintext credentials in code, config, or logs

**A05 — Injection**
- [ ] SQL: parameterized queries only (`?` placeholders)
- [ ] OS commands: no user input in `exec.Command`
- [ ] Path traversal: `filepath.Clean()` + prefix validation
- [ ] XSS: no `dangerouslySetInnerHTML`, metadata sanitized
- [ ] XML: `encoding/xml` marshaling, no string concatenation

**A07 — Authentication Failures**
- [ ] Login rate limiting
- [ ] Session tokens: httpOnly, Secure, SameSite=Strict
- [ ] API keys: sufficient entropy (32+ bytes from crypto/rand)

**A09 — Logging Failures**
- [ ] Auth events logged (login, logout, failures)
- [ ] No secrets in logs
- [ ] Request IDs for correlation

**A10 — Exception Handling**
- [ ] All errors handled (no unchecked returns)
- [ ] No panics in handlers
- [ ] External calls use context.WithTimeout

### 3. Go-Specific Checks

- [ ] `context.Context` propagated through all layers
- [ ] HTTP clients have timeout set
- [ ] File handles closed (via `defer`)
- [ ] `io.LimitReader` used for request bodies
- [ ] Race-free: no shared mutable state without sync

### 4. Report Format

For each finding:

```
**[SEVERITY]** Brief title
- File: path/to/file.go:42
- Issue: Description of the vulnerability
- Impact: What an attacker could achieve
- Fix: Specific code change recommended
```

Severities: CRITICAL, HIGH, MEDIUM, LOW, INFO
