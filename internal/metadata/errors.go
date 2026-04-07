package metadata

import "errors"

var (
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrRateLimited indicates the provider rate limit was exceeded.
	ErrRateLimited = errors.New("rate limited")

	// ErrProviderUnavailable indicates the provider service is unavailable.
	ErrProviderUnavailable = errors.New("provider unavailable")

	// ErrInvalidISBN indicates the provided ISBN is invalid.
	ErrInvalidISBN = errors.New("invalid ISBN")

	// ErrNoProviders indicates no providers are configured.
	ErrNoProviders = errors.New("no providers configured")
)
