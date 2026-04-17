# Proposal: Download Bypass Layer

**Status:** Draft  
**Date:** 2026-04-17  
**Problem:** All direct downloads via the Embedded Downloader fail because digital library sources (Anna's Archive, LibGen, Z-Library) protect download pages with Cloudflare challenges, DDoS-Guard, or session-gated redirects. A plain `http.Get` request receives a challenge page or HTTP 403/302 — never the file.

---

## Context

The current `internal/download/direct/download_file.go` makes an unauthenticated `HTTP GET` with only a `User-Agent: Bookaneer/1.0` header. Digital library sources return one of:

| Scenario | HTTP Response | Downloaded bytes |
|---|---|---|
| Cloudflare Turnstile/JS challenge | 200 (HTML challenge page) | ~50 KB HTML |
| DDoS-Guard | 200 (HTML) | ~20 KB HTML |
| Session-gated redirect | 302 → login page | 0 B |
| Geo-block | 403 | 0 B |

Shelfmark solves this with two bypass strategies:
1. **External**: [FlareSolverr](https://github.com/FlareSolverr/FlareSolverr) — an HTTP service that runs a headless browser, solves challenges, and returns the resolved HTML + cookies.
2. **Internal**: SeleniumBase/CDP — a headless Chromium instance that Shelfmark controls directly (heavier, richer).

Bookaneer is a Go binary targeting resource-constrained devices (NAS, Raspberry Pi). Embedding Chromium is not acceptable. The **external bypasser approach (FlareSolverr) fits the architecture**, as it is an optional sidecar service.

---

## Goals

1. Downloads from Anna's Archive, LibGen, and similar sources succeed when a bypass service is configured.
2. No mandatory external dependencies — bypass is opt-in via a new `bypass_clients` table (or a top-level config key).
3. The `Embedded Downloader` retries with bypass cookies/headers when the initial request fails with a challenge indicator.
4. Shelfmark's bypass API is compatible via a configurable URL — users who already run FlareSolverr with Shelfmark can reuse it.

---

## Non-Goals

- Embedding a headless browser in the Bookaneer binary.
- Bypassing CAPTCHAs that require human interaction.
- Bypassing paid-access or DRM-protected content.

---

## Architecture

```
                         ┌──────────────────────────────────┐
                         │        internal/bypass/          │
                         │                                  │
  download_file.go  ───► │  Resolver (detects challenge)    │
                         │  ├─ FlareSolverr client          │
                         │  └─ NoopBypasser (disabled)      │
                         └──────────────────────────────────┘
                                        │
                              ┌─────────┴─────────┐
                              │   FlareSolverr     │
                              │  (sidecar Docker)  │
                              └────────────────────┘
```

### New package: `internal/bypass/`

```
internal/bypass/
  bypass.go          — Bypasser interface + result type
  noop.go            — No-op implementation (bypass disabled)
  flaresolverr/
    client.go        — FlareSolverr HTTP client
    client_test.go
  challenge/
    detect.go        — Challenge detection (Cloudflare, DDoS-Guard, redirect)
    detect_test.go
```

### Interface

```go
// package bypass

// Result contains cookies and headers obtained after solving a challenge.
type Result struct {
    Cookies []*http.Cookie
    Headers map[string]string
    // ResolvedURL is the final URL after redirects.
    ResolvedURL string
}

// Bypasser resolves anti-bot challenges and returns usable credentials.
type Bypasser interface {
    // Solve navigates to url, solves any challenge, and returns session data.
    // Returns ErrChallengeUnsolvable if the challenge cannot be handled.
    Solve(ctx context.Context, url string) (*Result, error)
    // Enabled reports whether the bypasser is configured and active.
    Enabled() bool
}
```

### Challenge detection

```go
// package bypass/challenge

// indicators that identify a page as a bot-challenge gate
var cloudflareIndicators = []string{
    "just a moment",
    "verify you are human",
    "cloudflare.com/products/turnstile",
}

var ddosGuardIndicators = []string{
    "ddos-guard",
    "checking your browser before accessing",
}

// IsChallengePage returns true when body content matches known challenge patterns.
// body should be the first 4 KB of the response body.
func IsChallengePage(statusCode int, contentType, body string) bool
```

### Integration into `download_file.go`

The download flow becomes a two-attempt strategy:

```
Attempt 1: plain GET
  ├─ success (200 + Content-Type != text/html + size > threshold) → write file
  └─ challenge detected (HTML body or 302/403)
       ├─ bypasser disabled → mark failed with explicit message
       └─ bypasser enabled  → call Solve(url) → retry GET with cookies+headers
            ├─ success → write file
            └─ still challenged → mark failed with "challenge unsolvable"
```

### Configuration

New row type in the existing `config` table (key/value store), **or** a dedicated `bypass_clients` table if multiple bypass configurations are needed in the future:

```sql
-- migration 018_bypass_config.sql
CREATE TABLE bypass_clients (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT    NOT NULL DEFAULT 'FlareSolverr',
    type         TEXT    NOT NULL DEFAULT 'flaresolverr',  -- flaresolverr | none
    url          TEXT    NOT NULL DEFAULT '',
    timeout_ms   INTEGER NOT NULL DEFAULT 60000,
    enabled      INTEGER NOT NULL DEFAULT 0
);
```

Backend `ClientConfig`-style struct + repository (mirrors existing `download_clients` pattern).

Frontend: new card in **Settings → Download** with URL input + Test button — same shape as existing download client cards.

---

## Implementation Plan

### Phase 1 — Challenge detection (no external deps)

**Goal:** give users accurate error messages immediately.

1. `internal/bypass/challenge/detect.go` — `IsChallengePage(statusCode int, contentType, bodyPreview string) bool`
2. Update `download_file.go` to read the first 4 KB of the response, call `IsChallengePage`, and set `errorMessage` to a human-readable string:
   - `"Cloudflare challenge — configure a bypass service in Settings"`
   - `"DDoS-Guard challenge — configure a bypass service in Settings"`
   - `"HTTP {status}: {statusText}"`
3. Tests for `detect.go`.

**Files changed:** `internal/bypass/challenge/detect.go`, `internal/download/direct/download_file.go`  
**No new tables required.**

---

### Phase 2 — FlareSolverr client

**Goal:** implement the bypass interface against FlareSolverr v2 API.

1. `internal/bypass/bypass.go` — `Bypasser` interface + `Result` type + sentinel errors.
2. `internal/bypass/noop.go` — no-op implementation.
3. `internal/bypass/flaresolverr/client.go`:
   - `POST /v1` with `{"cmd": "request.get", "url": "...", "maxTimeout": 60000}`
   - Parse `solution.cookies` and `solution.response`
   - Return `bypass.Result{Cookies, Headers{"User-Agent": solution.userAgent}}`
4. Unit tests with `httptest.Server` mocking FlareSolverr responses.

**FlareSolverr v2 request/response contract:**
```json
// Request
{"cmd": "request.get", "url": "https://annas-archive.org/...", "maxTimeout": 60000}

// Response
{
  "status": "ok",
  "solution": {
    "url": "https://annas-archive.org/...",
    "status": 200,
    "headers": {...},
    "response": "<html>...",
    "cookies": [{"name": "cf_clearance", "value": "...", ...}],
    "userAgent": "Mozilla/5.0 ..."
  }
}
```

---

### Phase 3 — Wiring: config table + service + download retry

1. Migration `018_bypass_clients.sql` — `bypass_clients` table (schema above).
2. `internal/bypass/repository.go` — CRUD mirroring `internal/download/repository.go`.
3. `internal/bypass/service.go` — `GetBypasser(ctx) (Bypasser, error)` returns configured client or noop.
4. Wire `bypass.Service` into `download.Service` (constructor injection, same pattern as existing deps).
5. Update `download_file.go`:
   - After a challenge is detected call `bypasser.Solve(ctx, url)`.
   - Retry `http.Get` with cookies + `User-Agent` header from result.
   - On second failure, set `errorMessage` explaining bypass also failed.

---

### Phase 4 — Frontend: bypass client settings UI

1. New API handler `GET/POST/PUT/DELETE /api/v1/bypass-clients` (mirrors `download.go` handler pattern).
2. Frontend:
   - Zod schema `BypassClientSchema` in `web/src/lib/schemas/bypass.schema.ts`
   - API client `web/src/lib/api/bypass.ts`
   - Settings card in `web/src/containers/settings/BypassSettings.tsx`
   - Test button that calls `POST /api/v1/bypass-clients/{id}/test`
3. The settings page shows a **status indicator**: "FlareSolverr: connected / unreachable / disabled"

---

### Phase 5 — Docker Compose integration (optional sidecar)

Update `docker-compose.yml` and `docker-compose.dev.yml` with an **optional** FlareSolverr service, **commented out by default**:

```yaml
# Uncomment to enable Cloudflare bypass (requires more RAM)
# flaresolverr:
#   image: ghcr.io/flaresolverr/flaresolverr:latest
#   container_name: bookaneer-flaresolverr
#   environment:
#     - LOG_LEVEL=info
#   restart: unless-stopped
```

---

## Risk & Mitigations

| Risk | Mitigation |
|---|---|
| FlareSolverr is unmaintained / Cloudflare changes detection | Interface is abstract — alternative bypass backends can be added without touching download code |
| FlareSolverr requires ~1 GB RAM (Chromium) | It's opt-in; lightweight devices skip it |
| Anna's Archive DMCA / policy change | Bypass only affects HTTP transport; no scraping logic changes |
| False positives in challenge detection | Detection is read-only — worst case is an unnecessary bypass call that succeeds or fails gracefully |

---

## Open Questions

1. Should `bypass_clients` support multiple entries (e.g., one per source domain), or is a single global bypasser sufficient for v1?
2. Should Phase 3 also attempt bypass for non-embedded (qBittorrent, SABnzbd) clients that receive magnet/NZB links from challenged indexer pages? (Likely no — indexer challenges are separate.)
3. FlareSolverr v2 may not be needed for all sources — LibGen typically does not use Cloudflare. Should challenge detection skip the bypass call for non-challenge failures (4xx) to avoid latency?

---

## Acceptance Criteria

- [ ] A download from Anna's Archive succeeds end-to-end when FlareSolverr is configured.
- [ ] With bypass disabled, failed downloads show `"Cloudflare challenge — configure a bypass service in Settings"` instead of a generic error.
- [ ] `bypass.Bypasser` interface has 100% test coverage via `httptest`.
- [ ] The `bypass_clients` settings card has a working Test button.
- [ ] No new mandatory Docker dependencies introduced.
- [ ] All existing tests continue to pass.
