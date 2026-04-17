package challenge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woliveiras/bookaneer/internal/bypass/challenge"
)

func TestDetect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		body       string
		wantFound  bool
		wantReason string
	}{
		{
			name:       "cloudflare just a moment",
			body:       `<html><head><title>Just a moment...</title></head><body>Verify you are human</body></html>`,
			wantFound:  true,
			wantReason: "cloudflare",
		},
		{
			name:       "cloudflare turnstile indicator",
			body:       `<html><body><div id="challenge-platform">cloudflare.com/products/turnstile</div></body></html>`,
			wantFound:  true,
			wantReason: "cloudflare",
		},
		{
			name:       "ddos-guard",
			body:       `<!DOCTYPE html><html><body>DDoS-Guard is checking your browser before accessing the site.</body></html>`,
			wantFound:  true,
			wantReason: "ddosguard",
		},
		{
			name:       "login wall",
			body:       `<html><body><h1>Sign in to continue</h1></body></html>`,
			wantFound:  true,
			wantReason: "login",
		},
		{
			name:       "login required",
			body:       `<html><body>Login required to access this file.</body></html>`,
			wantFound:  true,
			wantReason: "login",
		},
		{
			name:      "normal epub binary header",
			body:      "PK\x03\x04 some epub content here",
			wantFound: false,
		},
		{
			name:      "normal pdf binary header",
			body:      "%PDF-1.4 some pdf content here",
			wantFound: false,
		},
		{
			name:      "empty body",
			body:      "",
			wantFound: false,
		},
		{
			name:      "normal html page not a challenge",
			body:      "<html><body><h1>Book Library</h1><p>Download your book here.</p></body></html>",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			found, reason := challenge.Detect(tt.body)
			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.wantReason, reason)
			}
		})
	}
}

func TestIsHTML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		contentType string
		want        bool
	}{
		{"text/html; charset=utf-8", true},
		{"TEXT/HTML", true},
		{"text/html", true},
		{"application/epub+zip", false},
		{"application/pdf", false},
		{"application/octet-stream", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, challenge.IsHTML(tt.contentType))
		})
	}
}
