// Package challenge provides challenge-page detection for common anti-bot systems.
package challenge

import "strings"

// cloudflareIndicators are lowercase phrases that identify Cloudflare challenge pages.
var cloudflareIndicators = []string{
	"just a moment",
	"verify you are human",
	"verifying you are human",
	"cloudflare.com/products/turnstile",
	"cf-browser-verification",
	"challenge-platform",
}

// ddosGuardIndicators are lowercase phrases that identify DDoS-Guard challenge pages.
var ddosGuardIndicators = []string{
	"ddos-guard",
	"ddos guard",
	"checking your browser before accessing",
	"complete the manual check to continue",
	"could not verify your browser automatically",
}

// loginIndicators are lowercase phrases that identify a login-wall interstitial.
var loginIndicators = []string{
	"sign in to continue",
	"log in to continue",
	"login required",
	"please log in",
	"you must be logged in",
}

// IsHTML returns true when the Content-Type header indicates an HTML document.
func IsHTML(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/html")
}

// Detect checks a body preview (first few KB is sufficient) for known challenge
// indicators. It returns (true, reason) when a challenge is detected where reason
// is one of "cloudflare", "ddosguard", or "login".
func Detect(bodyPreview string) (bool, string) {
	lower := strings.ToLower(bodyPreview)
	for _, indicator := range cloudflareIndicators {
		if strings.Contains(lower, indicator) {
			return true, "cloudflare"
		}
	}
	for _, indicator := range ddosGuardIndicators {
		if strings.Contains(lower, indicator) {
			return true, "ddosguard"
		}
	}
	for _, indicator := range loginIndicators {
		if strings.Contains(lower, indicator) {
			return true, "login"
		}
	}
	return false, ""
}
