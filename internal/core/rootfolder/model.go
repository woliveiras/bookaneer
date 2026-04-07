package rootfolder

// RootFolder represents a root folder for the library.
type RootFolder struct {
	ID                      int64  `json:"id"`
	Path                    string `json:"path"`
	Name                    string `json:"name"`
	DefaultQualityProfileID *int64 `json:"defaultQualityProfileId,omitempty"`

	// Computed fields
	FreeSpace   int64 `json:"freeSpace,omitempty"`
	TotalSpace  int64 `json:"totalSpace,omitempty"`
	AuthorCount int   `json:"authorCount,omitempty"`
	Accessible  bool  `json:"accessible"`
}

// CreateRootFolderInput holds the data needed to create a new root folder.
type CreateRootFolderInput struct {
	Path                    string `json:"path"`
	Name                    string `json:"name"`
	DefaultQualityProfileID *int64 `json:"defaultQualityProfileId,omitempty"`
}

// UpdateRootFolderInput holds the data for updating an existing root folder.
type UpdateRootFolderInput struct {
	Path                    *string `json:"path,omitempty"`
	Name                    *string `json:"name,omitempty"`
	DefaultQualityProfileID *int64  `json:"defaultQualityProfileId,omitempty"`
}
