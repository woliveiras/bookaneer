package series

// Series represents a book series.
type Series struct {
	ID          int64  `json:"id"`
	ForeignID   string `json:"foreignId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Monitored   bool   `json:"monitored"`
	BookCount   int    `json:"bookCount,omitempty"`
}

// SeriesBook represents a book in a series with its position.
type SeriesBook struct {
	SeriesID   int64  `json:"seriesId"`
	BookID     int64  `json:"bookId"`
	Position   string `json:"position"`
	BookTitle  string `json:"bookTitle,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
}

// CreateSeriesInput holds the data needed to create a new series.
type CreateSeriesInput struct {
	ForeignID   string `json:"foreignId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Monitored   bool   `json:"monitored"`
}

// UpdateSeriesInput holds the data for updating an existing series.
type UpdateSeriesInput struct {
	ForeignID   *string `json:"foreignId,omitempty"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Monitored   *bool   `json:"monitored,omitempty"`
}

// AddBookInput holds the data for adding a book to a series.
type AddBookInput struct {
	BookID   int64  `json:"bookId"`
	Position string `json:"position"`
}

// ListSeriesFilter provides filtering options for listing series.
type ListSeriesFilter struct {
	Monitored *bool
	Search    string
	SortBy    string
	SortDir   string
	Limit     int
	Offset    int
}

// SeriesWithBooks represents a series with its books.
type SeriesWithBooks struct {
	Series
	Books []SeriesBook `json:"books"`
}
