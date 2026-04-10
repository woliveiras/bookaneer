package search

import "errors"

var (
	// ErrIndexerNotFound is returned when an indexer is not found.
	ErrIndexerNotFound = errors.New("indexer not found")
	// ErrInvalidAPIKey is returned when the API key is invalid.
	ErrInvalidAPIKey = errors.New("invalid API key")
	// ErrRateLimited is returned when the indexer rate limits requests.
	ErrRateLimited = errors.New("rate limited")
	// ErrUnsupportedSearch is returned when the search type is unsupported.
	ErrUnsupportedSearch = errors.New("unsupported search type")
	// ErrInvalidResponse is returned when the indexer returns an invalid response.
	ErrInvalidResponse = errors.New("invalid response from indexer")
)
