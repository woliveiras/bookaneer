// Package bypass provides an abstraction for resolving anti-bot challenges
// (Cloudflare Turnstile, DDoS-Guard) encountered during direct file downloads.
// Concrete implementations live in sub-packages (e.g. bypass/flaresolverr).
package bypass

import (
	"context"
	"errors"
	"net/http"
)

// ErrDisabled is returned by the no-op Bypasser when no service is configured.
var ErrDisabled = errors.New("bypass service not configured")

// ErrUnsolvable is returned when the external service cannot solve the challenge.
var ErrUnsolvable = errors.New("challenge could not be solved")

// Result contains the session credentials obtained after solving a challenge.
// These are applied to the subsequent file-download request.
type Result struct {
	// Cookies are session cookies (e.g. cf_clearance) to attach to the retry.
	Cookies []*http.Cookie
	// UserAgent is the browser user-agent string reported by the bypass service.
	// Using the same UA as the challenge solution avoids fingerprint mismatches.
	UserAgent string
}

// Bypasser resolves anti-bot challenges and returns usable session credentials.
// The interface is defined here (consumer side) and implemented in sub-packages.
type Bypasser interface {
	// Solve navigates to url, resolves any challenge, and returns session data.
	// Returns ErrUnsolvable if the challenge cannot be handled.
	Solve(ctx context.Context, url string) (*Result, error)
	// Enabled reports whether the bypasser is configured and active.
	// A disabled Bypasser always returns ErrDisabled from Solve.
	Enabled() bool
}
