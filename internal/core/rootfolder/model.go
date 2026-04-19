package rootfolder

// RootFolder represents a root folder for the library.
type RootFolder struct {
	ID                      int64  `json:"id" db:"id"`
	Path                    string `json:"path" db:"path"`
	Name                    string `json:"name" db:"name"`
	DefaultQualityProfileID *int64 `json:"defaultQualityProfileId,omitempty" db:"default_quality_profile_id"`

	// Computed fields (not stored in DB)
	FreeSpace   int64 `json:"freeSpace,omitempty" db:"-"`
	TotalSpace  int64 `json:"totalSpace,omitempty" db:"-"`
	AuthorCount int   `json:"authorCount,omitempty" db:"-"`
	Accessible  bool  `json:"accessible" db:"-"`
}

// CreateRootFolderInput holds the data needed to create a new root folder.
type CreateRootFolderInput struct {
	Path                    string `json:"path" db:"path"`
	Name                    string `json:"name" db:"name"`
	DefaultQualityProfileID *int64 `json:"defaultQualityProfileId,omitempty" db:"default_quality_profile_id"`
}

// UpdateRootFolderInput holds the data for updating an existing root folder.
type UpdateRootFolderInput struct {
	Path                    *string `json:"path,omitempty"`
	Name                    *string `json:"name,omitempty"`
	DefaultQualityProfileID *int64  `json:"defaultQualityProfileId,omitempty"`
	MoveFiles               bool    `json:"moveFiles,omitempty"` // If true, move existing files to new path
}

// MigrationProgress represents the progress of a root folder migration.
type MigrationProgress struct {
	RootFolderID int64  `json:"rootFolderId"`
	OldPath      string `json:"oldPath"`
	NewPath      string `json:"newPath"`
	TotalAuthors int    `json:"totalAuthors"`
	MovedAuthors int    `json:"movedAuthors"`
	TotalFiles   int    `json:"totalFiles"`
	MovedFiles   int    `json:"movedFiles"`
	Status       string `json:"status"` // pending, in_progress, completed, failed
	Error        string `json:"error,omitempty"`
}
