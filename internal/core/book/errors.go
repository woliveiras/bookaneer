package book

import "errors"

var (
	// ErrNotFound is returned when a book is not found.
	ErrNotFound = errors.New("book not found")

	// ErrEditionNotFound is returned when an edition is not found.
	ErrEditionNotFound = errors.New("edition not found")

	// ErrFileNotFound is returned when a book file is not found.
	ErrFileNotFound = errors.New("book file not found")

	// ErrDuplicate is returned when a book with the same foreign ID already exists.
	ErrDuplicate = errors.New("book already exists")

	// ErrInvalidInput is returned when the input data is invalid.
	ErrInvalidInput = errors.New("invalid book input")

	// ErrAuthorNotFound is returned when the referenced author does not exist.
	ErrAuthorNotFound = errors.New("author not found")
)
