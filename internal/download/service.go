package download

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Service manages download clients and queue operations.
type Service struct {
	db                   *sql.DB
	mu                   sync.RWMutex
	clients              map[int64]Client
	embeddedClient       Client        // Auto-configured direct downloader
	embeddedClientConfig *ClientConfig // Cached config with DownloadDir
}

// NewService creates a new download service.
func NewService(db *sql.DB) *Service {
	return &Service{
		db:      db,
		clients: make(map[int64]Client),
	}
}

// TestClient tests connection to a download client.
func (s *Service) TestClient(ctx context.Context, cfg *ClientConfig) error {
	client, err := NewClient(*cfg)
	if err != nil {
		return err
	}
	return client.Test(ctx)
}

// GetQueue returns the combined queue from all enabled clients.
func (s *Service) GetQueue(ctx context.Context) ([]ItemStatus, error) {
	clients, err := s.ListClients(ctx)
	if err != nil {
		return nil, err
	}

	var allItems []ItemStatus
	for _, cfg := range clients {
		if !cfg.Enabled {
			continue
		}

		client, err := s.getOrCreateClient(cfg)
		if err != nil {
			continue
		}

		items, err := client.GetQueue(ctx)
		if err != nil {
			continue
		}

		allItems = append(allItems, items...)
	}

	return allItems, nil
}

// GetClientQueue returns the queue from a specific client.
func (s *Service) GetClientQueue(ctx context.Context, clientID int64) ([]ItemStatus, error) {
	cfg, err := s.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	client, err := s.getOrCreateClient(*cfg)
	if err != nil {
		return nil, err
	}

	return client.GetQueue(ctx)
}

// ListGrabs returns all grabs.
func (s *Service) ListGrabs(ctx context.Context) ([]GrabItem, error) {
	query := `
		SELECT id, book_id, indexer_id, release_title, download_url, size, quality,
		       client_id, download_id, status, error_message, grabbed_at, completed_at
		FROM grabs
		ORDER BY grabbed_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query grabs: %w", err)
	}
	defer rows.Close()

	var grabs []GrabItem
	for rows.Next() {
		var g GrabItem
		var clientID, downloadID sql.NullInt64
		var downloadIDStr sql.NullString
		var errorMsg sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&g.ID, &g.BookID, &g.IndexerID, &g.ReleaseTitle, &g.DownloadURL, &g.Size,
			&g.Quality, &clientID, &downloadIDStr, &g.Status, &errorMsg, &g.GrabbedAt, &completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan grab: %w", err)
		}

		if clientID.Valid {
			g.ClientID = clientID.Int64
		}
		if downloadIDStr.Valid {
			g.DownloadID = downloadIDStr.String
		} else if downloadID.Valid {
			g.DownloadID = fmt.Sprintf("%d", downloadID.Int64)
		}
		g.ErrorMessage = errorMsg.String
		if completedAt.Valid {
			g.CompletedAt = &completedAt.Time
		}

		grabs = append(grabs, g)
	}

	return grabs, rows.Err()
}

// CreateGrab creates a new grab record.
func (s *Service) CreateGrab(ctx context.Context, grab *GrabItem) error {
	now := time.Now().UTC()
	grab.GrabbedAt = now
	grab.Status = GrabStatusPending

	query := `
		INSERT INTO grabs (book_id, indexer_id, release_title, download_url, size, quality,
		                   status, grabbed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		grab.BookID, grab.IndexerID, grab.ReleaseTitle, grab.DownloadURL,
		grab.Size, grab.Quality, grab.Status, now,
	)
	if err != nil {
		return fmt.Errorf("insert grab: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	grab.ID = id

	return nil
}

// SendGrab sends a grab to a download client.
func (s *Service) SendGrab(ctx context.Context, grabID int64, clientID int64) error {
	// Get the grab
	var grab GrabItem
	var errorMsg sql.NullString
	var completedAt sql.NullTime
	var clientIDNull, downloadIDNull sql.NullInt64

	query := `
		SELECT id, book_id, indexer_id, release_title, download_url, size, quality,
		       client_id, download_id, status, error_message, grabbed_at, completed_at
		FROM grabs WHERE id = ?
	`
	err := s.db.QueryRowContext(ctx, query, grabID).Scan(
		&grab.ID, &grab.BookID, &grab.IndexerID, &grab.ReleaseTitle, &grab.DownloadURL,
		&grab.Size, &grab.Quality, &clientIDNull, &downloadIDNull, &grab.Status, &errorMsg,
		&grab.GrabbedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("query grab: %w", err)
	}

	// Get the client
	cfg, err := s.GetClient(ctx, clientID)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return ErrClientDisabled
	}

	client, err := s.getOrCreateClient(*cfg)
	if err != nil {
		return err
	}

	// Add to client
	downloadID, err := client.Add(ctx, AddItem{
		Name:        grab.ReleaseTitle,
		DownloadURL: grab.DownloadURL,
		Category:    cfg.Category,
		Priority:    cfg.RecentPriority,
	})
	if err != nil {
		// Update grab with error
		_, _ = s.db.ExecContext(ctx,
			"UPDATE grabs SET status = ?, error_message = ? WHERE id = ?",
			GrabStatusFailed, err.Error(), grabID,
		)
		return fmt.Errorf("add to client: %w", err)
	}

	// Update grab with success
	_, err = s.db.ExecContext(ctx,
		"UPDATE grabs SET client_id = ?, download_id = ?, status = ? WHERE id = ?",
		clientID, downloadID, GrabStatusSent, grabID,
	)
	if err != nil {
		return fmt.Errorf("update grab: %w", err)
	}

	return nil
}

// GrabRequest is the request body for creating a grab.
type GrabRequest struct {
	BookID       int64  `json:"bookId"`
	IndexerID    int64  `json:"indexerId"`
	ReleaseTitle string `json:"releaseTitle"`
	DownloadURL  string `json:"downloadUrl"`
	Size         int64  `json:"size"`
	Quality      string `json:"quality"`
}

// UnmarshalJSON implements json.Unmarshaler for GrabItem.
func (g *GrabItem) UnmarshalJSON(data []byte) error {
	type Alias GrabItem
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(g),
	}
	return json.Unmarshal(data, &aux)
}
