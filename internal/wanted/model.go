package wanted

// GrabResult represents the result of a grab attempt.
type GrabResult struct {
	BookID       int64  `json:"bookId"`
	Title        string `json:"title"`
	Source       string `json:"source"` // "library" or "indexer"
	ProviderName string `json:"providerName"`
	Format       string `json:"format"`
	Size         int64  `json:"size"`
	DownloadID   string `json:"downloadId"` // ID from download client
	ClientName   string `json:"clientName"`
}

// DownloadQueueItem represents an item in the download queue.
type DownloadQueueItem struct {
	ID               int64   `json:"id"`
	BookID           int64   `json:"bookId"`
	DownloadClientID *int64  `json:"downloadClientId,omitempty"`
	IndexerID        *int64  `json:"indexerId,omitempty"`
	ExternalID       string  `json:"externalId"`
	Title            string  `json:"title"`
	Size             int64   `json:"size"`
	Format           string  `json:"format"`
	Status           string  `json:"status"`
	Progress         float64 `json:"progress"`
	DownloadURL      string  `json:"downloadUrl"`
	AddedAt          string  `json:"addedAt"`
	BookTitle        string  `json:"bookTitle"`
	ClientName       string  `json:"clientName"`
	ErrorMessage     string  `json:"errorMessage,omitempty"`
}

// HistoryItem represents a history event.
type HistoryItem struct {
	ID          int64          `json:"id"`
	BookID      *int64         `json:"bookId,omitempty"`
	AuthorID    *int64         `json:"authorId,omitempty"`
	EventType   string         `json:"eventType"`
	SourceTitle string         `json:"sourceTitle"`
	Quality     string         `json:"quality"`
	Data        map[string]any `json:"data"`
	Date        string         `json:"date"`
	BookTitle   string         `json:"bookTitle,omitempty"`
	AuthorName  string         `json:"authorName,omitempty"`
}

// BlocklistItem represents a blocked release.
type BlocklistItem struct {
	ID          int64  `json:"id"`
	BookID      int64  `json:"bookId"`
	SourceTitle string `json:"sourceTitle"`
	Quality     string `json:"quality"`
	Reason      string `json:"reason"`
	Date        string `json:"date"`
	BookTitle   string `json:"bookTitle"`
	AuthorName  string `json:"authorName"`
}
