package book

// Book represents a book in the library.
type Book struct {
	ID          int64  `json:"id"`
	AuthorID    int64  `json:"authorId"`
	Title       string `json:"title"`
	SortTitle   string `json:"sortTitle"`
	ForeignID   string `json:"foreignId"` // OpenLibrary work key
	ISBN        string `json:"isbn"`
	ISBN13      string `json:"isbn13"`
	ReleaseDate string `json:"releaseDate"` // YYYY-MM-DD
	Overview    string `json:"overview"`
	ImageURL    string `json:"imageUrl"`
	PageCount   int    `json:"pageCount"`
	Monitored   bool   `json:"monitored"`
	UserRating  *int   `json:"userRating,omitempty"` // 1-5 stars, nil = unrated
	InWishlist  bool   `json:"inWishlist"`
	AddedAt     string `json:"addedAt"`
	UpdatedAt   string `json:"updatedAt"`

	// Computed/joined fields
	AuthorName string `json:"authorName,omitempty"`
	HasFile    bool   `json:"hasFile,omitempty"`
	FileFormat string `json:"fileFormat,omitempty"` // format of the primary file on disk
}

// Edition represents a specific edition of a book.
type Edition struct {
	ID          int64  `json:"id"`
	BookID      int64  `json:"bookId"`
	ForeignID   string `json:"foreignId"` // OpenLibrary edition key
	Title       string `json:"title"`
	ISBN        string `json:"isbn"`
	ISBN13      string `json:"isbn13"`
	Format      string `json:"format"` // epub, mobi, pdf, hardcover, paperback
	Publisher   string `json:"publisher"`
	ReleaseDate string `json:"releaseDate"`
	PageCount   int    `json:"pageCount"`
	Language    string `json:"language"` // ISO 639-1
	Monitored   bool   `json:"monitored"`
}

// BookFile represents a book file on disk.
type BookFile struct {
	ID              int64  `json:"id"`
	BookID          int64  `json:"bookId"`
	EditionID       *int64 `json:"editionId,omitempty"`
	Path            string `json:"path"`
	RelativePath    string `json:"relativePath"`
	Size            int64  `json:"size"`
	Format          string `json:"format"` // epub, mobi, azw3, pdf, cbz
	Quality         string `json:"quality"`
	Hash            string `json:"hash"` // SHA-256
	AddedAt         string `json:"addedAt"`
	ContentMismatch bool   `json:"contentMismatch"`
}

// CreateBookInput holds the data needed to create a new book.
type CreateBookInput struct {
	AuthorID    int64  `json:"authorId"`
	Title       string `json:"title"`
	SortTitle   string `json:"sortTitle"`
	ForeignID   string `json:"foreignId"`
	ISBN        string `json:"isbn"`
	ISBN13      string `json:"isbn13"`
	ReleaseDate string `json:"releaseDate"`
	Overview    string `json:"overview"`
	ImageURL    string `json:"imageUrl"`
	PageCount   int    `json:"pageCount"`
	Monitored   bool   `json:"monitored"`
	InWishlist  bool   `json:"inWishlist"`
}

// UpdateBookInput holds the data for updating an existing book.
type UpdateBookInput struct {
	AuthorID    *int64  `json:"authorId,omitempty"`
	Title       *string `json:"title,omitempty"`
	SortTitle   *string `json:"sortTitle,omitempty"`
	ForeignID   *string `json:"foreignId,omitempty"`
	ISBN        *string `json:"isbn,omitempty"`
	ISBN13      *string `json:"isbn13,omitempty"`
	ReleaseDate *string `json:"releaseDate,omitempty"`
	Overview    *string `json:"overview,omitempty"`
	ImageURL    *string `json:"imageUrl,omitempty"`
	PageCount   *int    `json:"pageCount,omitempty"`
	Monitored   *bool   `json:"monitored,omitempty"`
	UserRating  *int    `json:"userRating,omitempty"` // 1-5, or 0 to clear
	InWishlist  *bool   `json:"inWishlist,omitempty"`
}

// CreateEditionInput holds the data needed to create a new edition.
type CreateEditionInput struct {
	BookID      int64  `json:"bookId"`
	ForeignID   string `json:"foreignId"`
	Title       string `json:"title"`
	ISBN        string `json:"isbn"`
	ISBN13      string `json:"isbn13"`
	Format      string `json:"format"`
	Publisher   string `json:"publisher"`
	ReleaseDate string `json:"releaseDate"`
	PageCount   int    `json:"pageCount"`
	Language    string `json:"language"`
	Monitored   bool   `json:"monitored"`
}

// ListBooksFilter provides filtering options for listing books.
type ListBooksFilter struct {
	AuthorID   *int64
	Monitored  *bool
	Missing    bool // Only books without files
	InWishlist bool // Only books in wishlist
	Search     string
	SortBy     string // title, sortTitle, releaseDate, addedAt, rating
	SortDir    string // asc, desc
	Limit      int
	Offset     int
}

// BookWithEditions represents a book with its editions.
type BookWithEditions struct {
	Book
	Editions []Edition  `json:"editions"`
	Files    []BookFile `json:"files,omitempty"`
}
