package rootfolder

import "errors"

var (
	// ErrNotFound is returned when a root folder is not found.
	ErrNotFound = errors.New("root folder not found")

	// ErrDuplicate is returned when a root folder with the same path already exists.
	ErrDuplicate = errors.New("root folder already exists")

	// ErrInvalidInput is returned when the input data is invalid.
	ErrInvalidInput = errors.New("invalid root folder input")

	// ErrPathNotAccessible is returned when the path is not accessible.
	ErrPathNotAccessible = errors.New("path is not accessible")

	// ErrMigrationFailed is returned when file migration fails.
	ErrMigrationFailed = errors.New("migration failed")
)
