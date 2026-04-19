package pathmapping

// RemotePathMapping maps a download client path to a local path.
type RemotePathMapping struct {
	ID         int64  `json:"id" db:"id"`
	Host       string `json:"host" db:"host"`              // download client host (informational)
	RemotePath string `json:"remotePath" db:"remote_path"` // path prefix as seen by the download client
	LocalPath  string `json:"localPath" db:"local_path"`   // corresponding local path prefix
	CreatedAt  string `json:"createdAt" db:"created_at"`
}

// CreateInput holds the data needed to create a new mapping.
type CreateInput struct {
	Host       string `json:"host" db:"host"`
	RemotePath string `json:"remotePath" db:"remote_path"`
	LocalPath  string `json:"localPath" db:"local_path"`
}

// UpdateInput holds the data for updating a mapping.
type UpdateInput struct {
	Host       *string `json:"host,omitempty"`
	RemotePath *string `json:"remotePath,omitempty"`
	LocalPath  *string `json:"localPath,omitempty"`
}
