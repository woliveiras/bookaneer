package pathmapping

// RemotePathMapping maps a download client path to a local path.
type RemotePathMapping struct {
	ID         int64  `json:"id"`
	Host       string `json:"host"`       // download client host (informational)
	RemotePath string `json:"remotePath"` // path prefix as seen by the download client
	LocalPath  string `json:"localPath"`  // corresponding local path prefix
	CreatedAt  string `json:"createdAt"`
}

// CreateInput holds the data needed to create a new mapping.
type CreateInput struct {
	Host       string `json:"host"`
	RemotePath string `json:"remotePath"`
	LocalPath  string `json:"localPath"`
}

// UpdateInput holds the data for updating a mapping.
type UpdateInput struct {
	Host       *string `json:"host,omitempty"`
	RemotePath *string `json:"remotePath,omitempty"`
	LocalPath  *string `json:"localPath,omitempty"`
}
