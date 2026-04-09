package download

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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

// ListClients returns all download clients from the database.
func (s *Service) ListClients(ctx context.Context) ([]ClientConfig, error) {
	query := `
		SELECT id, name, type, host, port, use_tls, username, password, api_key, 
		       category, recent_priority, older_priority, remove_completed_after, 
		       enabled, priority, nzb_folder, torrent_folder, watch_folder, download_dir,
		       created_at, updated_at
		FROM download_clients
		ORDER BY priority ASC, name ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query clients: %w", err)
	}
	defer rows.Close()

	var clients []ClientConfig
	for rows.Next() {
		var cfg ClientConfig
		var username, password, apiKey, category sql.NullString
		var nzbFolder, torrentFolder, watchFolder, downloadDir sql.NullString

		err := rows.Scan(
			&cfg.ID, &cfg.Name, &cfg.Type, &cfg.Host, &cfg.Port, &cfg.UseTLS,
			&username, &password, &apiKey, &category,
			&cfg.RecentPriority, &cfg.OlderPriority, &cfg.RemoveCompletedAfter,
			&cfg.Enabled, &cfg.Priority, &nzbFolder, &torrentFolder, &watchFolder, &downloadDir,
			&cfg.CreatedAt, &cfg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
		}

		cfg.Username = username.String
		cfg.Password = password.String
		cfg.APIKey = apiKey.String
		cfg.Category = category.String
		cfg.NzbFolder = nzbFolder.String
		cfg.TorrentFolder = torrentFolder.String
		cfg.WatchFolder = watchFolder.String
		cfg.DownloadDir = downloadDir.String

		clients = append(clients, cfg)
	}

	return clients, rows.Err()
}

// GetClient returns a download client by ID.
func (s *Service) GetClient(ctx context.Context, id int64) (*ClientConfig, error) {
	query := `
		SELECT id, name, type, host, port, use_tls, username, password, api_key, 
		       category, recent_priority, older_priority, remove_completed_after, 
		       enabled, priority, nzb_folder, torrent_folder, watch_folder, download_dir,
		       created_at, updated_at
		FROM download_clients
		WHERE id = ?
	`

	var cfg ClientConfig
	var username, password, apiKey, category sql.NullString
	var nzbFolder, torrentFolder, watchFolder, downloadDir sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&cfg.ID, &cfg.Name, &cfg.Type, &cfg.Host, &cfg.Port, &cfg.UseTLS,
		&username, &password, &apiKey, &category,
		&cfg.RecentPriority, &cfg.OlderPriority, &cfg.RemoveCompletedAfter,
		&cfg.Enabled, &cfg.Priority, &nzbFolder, &torrentFolder, &watchFolder, &downloadDir,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query client: %w", err)
	}

	cfg.Username = username.String
	cfg.Password = password.String
	cfg.APIKey = apiKey.String
	cfg.Category = category.String
	cfg.NzbFolder = nzbFolder.String
	cfg.TorrentFolder = torrentFolder.String
	cfg.WatchFolder = watchFolder.String
	cfg.DownloadDir = downloadDir.String

	return &cfg, nil
}

// CreateClient creates a new download client.
func (s *Service) CreateClient(ctx context.Context, cfg *ClientConfig) error {
	now := time.Now().UTC().Format(time.RFC3339)

	query := `
		INSERT INTO download_clients (
			name, type, host, port, use_tls, username, password, api_key,
			category, recent_priority, older_priority, remove_completed_after,
			enabled, priority, nzb_folder, torrent_folder, watch_folder, download_dir,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query,
		cfg.Name, cfg.Type, cfg.Host, cfg.Port, cfg.UseTLS,
		cfg.Username, cfg.Password, cfg.APIKey,
		cfg.Category, cfg.RecentPriority, cfg.OlderPriority, cfg.RemoveCompletedAfter,
		cfg.Enabled, cfg.Priority, cfg.NzbFolder, cfg.TorrentFolder,
		cfg.WatchFolder, cfg.DownloadDir, now, now,
	)
	if err != nil {
		return fmt.Errorf("insert client: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	cfg.ID = id
	cfg.CreatedAt = now
	cfg.UpdatedAt = now

	return nil
}

// UpdateClient updates an existing download client.
func (s *Service) UpdateClient(ctx context.Context, cfg *ClientConfig) error {
	now := time.Now().UTC().Format(time.RFC3339)

	query := `
		UPDATE download_clients SET
			name = ?, type = ?, host = ?, port = ?, use_tls = ?,
			username = ?, password = ?, api_key = ?, category = ?,
			recent_priority = ?, older_priority = ?, remove_completed_after = ?,
			enabled = ?, priority = ?, nzb_folder = ?, torrent_folder = ?,
			watch_folder = ?, download_dir = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := s.db.ExecContext(ctx, query,
		cfg.Name, cfg.Type, cfg.Host, cfg.Port, cfg.UseTLS,
		cfg.Username, cfg.Password, cfg.APIKey,
		cfg.Category, cfg.RecentPriority, cfg.OlderPriority, cfg.RemoveCompletedAfter,
		cfg.Enabled, cfg.Priority, cfg.NzbFolder, cfg.TorrentFolder,
		cfg.WatchFolder, cfg.DownloadDir, now, cfg.ID,
	)
	if err != nil {
		return fmt.Errorf("update client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	cfg.UpdatedAt = now

	// Invalidate cached client
	s.mu.Lock()
	delete(s.clients, cfg.ID)
	s.mu.Unlock()

	return nil
}

// DeleteClient deletes a download client.
func (s *Service) DeleteClient(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM download_clients WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()

	return nil
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

func (s *Service) getOrCreateClient(cfg ClientConfig) (Client, error) {
	s.mu.RLock()
	if client, ok := s.clients[cfg.ID]; ok {
		s.mu.RUnlock()
		return client, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check
	if client, ok := s.clients[cfg.ID]; ok {
		return client, nil
	}

	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	s.clients[cfg.ID] = client
	return client, nil
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

// GetDirectClient returns a direct download client for downloading
// from digital libraries (Internet Archive, LibGen, Anna's Archive).
// If no client is configured, creates an embedded one using the first root folder.
func (s *Service) GetDirectClient(ctx context.Context) (Client, *ClientConfig, error) {
	// First try to get a configured direct client
	client, cfg, err := s.getClientByType(ctx, ClientTypeDirect)
	if err != nil {
		return nil, nil, err
	}
	if client != nil {
		return client, cfg, nil
	}

	// No configured client, use embedded downloader with root folder
	return s.getEmbeddedDirectClient(ctx)
}

// getEmbeddedDirectClient returns or creates the embedded direct downloader
// that uses the first root folder as download destination.
func (s *Service) getEmbeddedDirectClient(ctx context.Context) (Client, *ClientConfig, error) {
	// Fast path: return cached client without holding lock during DB query
	s.mu.RLock()
	if s.embeddedClient != nil && s.embeddedClientConfig != nil {
		client, cfg := s.embeddedClient, s.embeddedClientConfig
		s.mu.RUnlock()
		return client, cfg, nil
	}
	s.mu.RUnlock()

	// Get the first root folder BEFORE acquiring write lock
	var rootPath string
	err := s.db.QueryRowContext(ctx, `SELECT path FROM root_folders ORDER BY id LIMIT 1`).Scan(&rootPath)
	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("no root folder configured: please add a root folder in Settings")
	}
	if err != nil {
		return nil, nil, fmt.Errorf("get root folder: %w", err)
	}

	// Now acquire lock to create client
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check: another goroutine may have created it while we were querying
	if s.embeddedClient != nil && s.embeddedClientConfig != nil {
		return s.embeddedClient, s.embeddedClientConfig, nil
	}

	// Create embedded direct client
	client, err := NewClient(ClientConfig{
		Name:        "Embedded Downloader",
		Type:        ClientTypeDirect,
		DownloadDir: rootPath,
		Enabled:     true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create embedded client: %w", err)
	}

	s.embeddedClient = client
	s.embeddedClientConfig = &ClientConfig{
		ID:          0,
		Name:        "Embedded Downloader",
		Type:        ClientTypeDirect,
		DownloadDir: rootPath,
		Enabled:     true,
	}

	return client, s.embeddedClientConfig, nil
}

// GetUsenetClient returns a configured usenet download client (SABnzbd, NZBGet).
// Returns nil, nil if no usenet client is configured.
func (s *Service) GetUsenetClient(ctx context.Context) (Client, *ClientConfig, error) {
	return s.getClientByType(ctx, ClientTypeSABnzbd)
}

// GetTorrentClient returns a configured torrent download client
// (qBittorrent, Transmission, Deluge).
// Returns nil, nil if no torrent client is configured.
func (s *Service) GetTorrentClient(ctx context.Context) (Client, *ClientConfig, error) {
	return s.getClientByTypes(ctx, ClientTypeQBittorrent, ClientTypeTransmission)
}

// getClientByType finds and returns the first enabled client of the given type.
func (s *Service) getClientByType(ctx context.Context, clientType string) (Client, *ClientConfig, error) {
	return s.getClientByTypes(ctx, clientType)
}

// getClientByTypes finds and returns the first enabled client of any of the given types.
func (s *Service) getClientByTypes(ctx context.Context, clientTypes ...string) (Client, *ClientConfig, error) {
	if len(clientTypes) == 0 {
		return nil, nil, nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(clientTypes))
	args := make([]any, len(clientTypes))
	for i, ct := range clientTypes {
		placeholders[i] = "?"
		args[i] = ct
	}

	query := fmt.Sprintf(`
		SELECT id, name, type, host, port, use_tls, username, password, api_key, 
		       category, enabled, priority, nzb_folder, torrent_folder, watch_folder, download_dir
		FROM download_clients
		WHERE type IN (%s) AND enabled = 1
		ORDER BY priority ASC
		LIMIT 1
	`, strings.Join(placeholders, ", "))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("query clients: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil, nil // No client configured
	}

	var cfg ClientConfig
	var enabled, useTLS int
	var username, password, apiKey, category sql.NullString
	var nzbFolder, torrentFolder, watchFolder, downloadDir sql.NullString
	if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.Type, &cfg.Host, &cfg.Port, &useTLS,
		&username, &password, &apiKey, &category, &enabled, &cfg.Priority,
		&nzbFolder, &torrentFolder, &watchFolder, &downloadDir); err != nil {
		return nil, nil, fmt.Errorf("scan client: %w", err)
	}
	cfg.Enabled = enabled == 1
	cfg.UseTLS = useTLS == 1
	cfg.Username = username.String
	cfg.Password = password.String
	cfg.APIKey = apiKey.String
	cfg.Category = category.String
	cfg.NzbFolder = nzbFolder.String
	cfg.TorrentFolder = torrentFolder.String
	cfg.WatchFolder = watchFolder.String
	cfg.DownloadDir = downloadDir.String

	client, err := s.getOrCreateClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	return client, &cfg, nil
}
