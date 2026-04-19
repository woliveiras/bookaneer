package author

// Author represents a book author in the library.
type Author struct {
	ID        int64  `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	SortName  string `json:"sortName" db:"sort_name"`   // "Tolkien, J.R.R."
	ForeignID string `json:"foreignId" db:"foreign_id"` // OpenLibrary author key
	Overview  string `json:"overview" db:"overview"`
	ImageURL  string `json:"imageUrl" db:"image_url"`
	Status    string `json:"status" db:"status"` // active, paused, ended
	Monitored bool   `json:"monitored" db:"monitored"`
	Path      string `json:"path" db:"path"` // /library/J.R.R. Tolkien
	AddedAt   string `json:"addedAt" db:"added_at"`
	UpdatedAt string `json:"updatedAt" db:"updated_at"`

	// Computed fields (not stored in DB)
	BookCount     int `json:"bookCount,omitempty"`
	BookFileCount int `json:"bookFileCount,omitempty"`
}

// CreateAuthorInput holds the data needed to create a new author.
type CreateAuthorInput struct {
	Name      string `json:"name"`
	SortName  string `json:"sortName"`
	ForeignID string `json:"foreignId"`
	Overview  string `json:"overview"`
	ImageURL  string `json:"imageUrl"`
	Status    string `json:"status"`
	Monitored bool   `json:"monitored"`
	Path      string `json:"path"`
}

// UpdateAuthorInput holds the data for updating an existing author.
type UpdateAuthorInput struct {
	Name      *string `json:"name,omitempty"`
	SortName  *string `json:"sortName,omitempty"`
	ForeignID *string `json:"foreignId,omitempty"`
	Overview  *string `json:"overview,omitempty"`
	ImageURL  *string `json:"imageUrl,omitempty"`
	Status    *string `json:"status,omitempty"`
	Monitored *bool   `json:"monitored,omitempty"`
	Path      *string `json:"path,omitempty"`
}

// ListAuthorsFilter provides filtering options for listing authors.
type ListAuthorsFilter struct {
	Monitored *bool
	Status    string
	Search    string
	SortBy    string // name, sortName, addedAt
	SortDir   string // asc, desc
	Limit     int
	Offset    int
}

// AuthorStats holds statistics for an author.
type AuthorStats struct {
	BookCount      int `json:"bookCount" db:"book_count"`
	BookFileCount  int `json:"bookFileCount" db:"file_count"`
	MissingBooks   int `json:"missingBooks" db:"missing"`
	TotalSizeBytes int `json:"totalSizeBytes" db:"total_size"`
}
