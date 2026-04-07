package reader

import "errors"

var (
	ErrBookFileNotFound = errors.New("book file not found")
	ErrProgressNotFound = errors.New("reading progress not found")
)

// BookFile represents an ebook file on disk.
type BookFile struct {
	ID           int64  `json:"id"`
	BookID       int64  `json:"bookId"`
	EditionID    *int64 `json:"editionId,omitempty"`
	Path         string `json:"path"`
	RelativePath string `json:"relativePath"`
	Size         int64  `json:"size"`
	Format       string `json:"format"`
	Quality      string `json:"quality"`
	Hash         string `json:"hash,omitempty"`
	AddedAt      string `json:"addedAt"`

	// Joined fields
	BookTitle  string `json:"bookTitle,omitempty"`
	AuthorName string `json:"authorName,omitempty"`
	CoverURL   string `json:"coverUrl,omitempty"`
}

// ReadingProgress represents a user's reading position in a book.
type ReadingProgress struct {
	ID         int64   `json:"id"`
	BookFileID int64   `json:"bookFileId"`
	UserID     int64   `json:"userId"`
	Position   string  `json:"position"`   // EPUB CFI string
	Percentage float64 `json:"percentage"` // 0.0 to 1.0
	UpdatedAt  string  `json:"updatedAt"`
}
