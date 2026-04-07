package download

import "errors"

var (
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidType indicates an invalid client type.
	ErrInvalidType = errors.New("invalid client type")

	// ErrConnectionFailed indicates connection to the download client failed.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrAuthFailed indicates authentication to the download client failed.
	ErrAuthFailed = errors.New("authentication failed")

	// ErrClientDisabled indicates the client is disabled.
	ErrClientDisabled = errors.New("client is disabled")
)
