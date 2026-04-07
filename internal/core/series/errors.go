package series

import "errors"

var (
	// ErrNotFound is returned when a series is not found.
	ErrNotFound = errors.New("series not found")

	// ErrDuplicate is returned when a series with the same foreign ID already exists.
	ErrDuplicate = errors.New("series already exists")

	// ErrInvalidInput is returned when the input data is invalid.
	ErrInvalidInput = errors.New("invalid series input")

	// ErrBookNotFound is returned when the referenced book does not exist.
	ErrBookNotFound = errors.New("book not found")

	// ErrBookAlreadyInSeries is returned when the book is already in the series.
	ErrBookAlreadyInSeries = errors.New("book already in series")
)
