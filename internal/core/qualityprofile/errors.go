package qualityprofile

import "errors"

var (
	// ErrNotFound is returned when a quality profile is not found.
	ErrNotFound = errors.New("quality profile not found")

	// ErrInvalidInput is returned when the input data is invalid.
	ErrInvalidInput = errors.New("invalid quality profile input")

	// ErrInUse is returned when trying to delete a profile that is in use.
	ErrInUse = errors.New("quality profile is in use")
)
