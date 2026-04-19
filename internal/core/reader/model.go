package reader

import "errors"

var (
	ErrBookFileNotFound = errors.New("book file not found")
	ErrProgressNotFound = errors.New("reading progress not found")
	ErrBookmarkNotFound = errors.New("bookmark not found")
)

// BookFile represents an ebook file on disk.
type BookFile struct {
	ID           int64  `json:"id" db:"id"`
	BookID       int64  `json:"bookId" db:"book_id"`
	EditionID    *int64 `json:"editionId,omitempty" db:"edition_id"`
	Path         string `json:"path" db:"path"`
	RelativePath string `json:"relativePath" db:"relative_path"`
	Size         int64  `json:"size" db:"size"`
	Format       string `json:"format" db:"format"`
	Quality      string `json:"quality" db:"quality"`
	Hash         string `json:"hash,omitempty" db:"hash"`
	AddedAt      string `json:"addedAt" db:"added_at"`

	// Joined fields
	BookTitle  string `json:"bookTitle,omitempty" db:"book_title"`
	AuthorName string `json:"authorName,omitempty" db:"author_name"`
	CoverURL   string `json:"coverUrl,omitempty" db:"cover_url"`
}

// ReadingProgress represents a user's reading position in a book.
type ReadingProgress struct {
	ID         int64   `json:"id" db:"id"`
	BookFileID int64   `json:"bookFileId" db:"book_file_id"`
	UserID     int64   `json:"userId" db:"user_id"`
	Position   string  `json:"position" db:"position"`     // EPUB CFI string
	Percentage float64 `json:"percentage" db:"percentage"` // 0.0 to 1.0
	UpdatedAt  string  `json:"updatedAt" db:"updated_at"`
}

// Bookmark represents a user's saved position in a book.
type Bookmark struct {
	ID         int64  `json:"id" db:"id"`
	BookFileID int64  `json:"bookFileId" db:"book_file_id"`
	UserID     int64  `json:"userId" db:"user_id"`
	Position   string `json:"position" db:"position"` // EPUB CFI or page number
	Title      string `json:"title" db:"title"`       // User-provided or auto-generated
	Note       string `json:"note" db:"note"`         // Optional user note
	CreatedAt  string `json:"createdAt" db:"created_at"`
}
