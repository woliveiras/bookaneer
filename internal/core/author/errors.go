package author

import "errors"

var (
	// ErrNotFound is returned when an author is not found.
	ErrNotFound = errors.New("author not found")

	// ErrDuplicate is returned when an author with the same foreign ID already exists.
	ErrDuplicate = errors.New("author already exists")

	// ErrInvalidInput is returned when the input data is invalid.
	ErrInvalidInput = errors.New("invalid author input")
)
