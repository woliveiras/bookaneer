package download

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/woliveiras/bookaneer/internal/database"
)

// clientScanner is an alias for database.Scanner (abstracts sql.Row and sql.Rows).
type clientScanner = database.Scanner

// scanClientConfig scans a full download_clients row (21 columns) into a ClientConfig.
// Column order: id, name, type, host, port, use_tls, username, password, api_key,
// category, recent_priority, older_priority, remove_completed_after,
// enabled, priority, nzb_folder, torrent_folder, watch_folder, download_dir,
// created_at, updated_at
func scanClientConfig(s clientScanner) (ClientConfig, error) {
	var cfg ClientConfig
	var username, password, apiKey, category sql.NullString
	var nzbFolder, torrentFolder, watchFolder, downloadDir sql.NullString

	err := s.Scan(
		&cfg.ID, &cfg.Name, &cfg.Type, &cfg.Host, &cfg.Port, &cfg.UseTLS,
		&username, &password, &apiKey, &category,
		&cfg.RecentPriority, &cfg.OlderPriority, &cfg.RemoveCompletedAfter,
		&cfg.Enabled, &cfg.Priority, &nzbFolder, &torrentFolder, &watchFolder, &downloadDir,
		&cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return cfg, err
	}

	cfg.Username = username.String
	cfg.Password = password.String
	cfg.APIKey = apiKey.String
	cfg.Category = category.String
	cfg.NzbFolder = nzbFolder.String
	cfg.TorrentFolder = torrentFolder.String
	cfg.WatchFolder = watchFolder.String
	cfg.DownloadDir = downloadDir.String

	return cfg, nil
}

const clientSelectColumns = `
	SELECT id, name, type, host, port, use_tls, username, password, api_key, 
	       category, recent_priority, older_priority, remove_completed_after, 
	       enabled, priority, nzb_folder, torrent_folder, watch_folder, download_dir,
	       created_at, updated_at
	FROM download_clients
`

// ListClients returns all download clients from the database.
func (s *Service) ListClients(ctx context.Context) ([]ClientConfig, error) {
	rows, err := s.db.QueryContext(ctx, clientSelectColumns+`ORDER BY priority ASC, name ASC`)
	if err != nil {
		return nil, fmt.Errorf("query clients: %w", err)
	}
	defer rows.Close()

	var clients []ClientConfig
	for rows.Next() {
		cfg, err := scanClientConfig(rows)
		if err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
		}
		clients = append(clients, cfg)
	}

	return clients, rows.Err()
}

// GetClient returns a download client by ID.
func (s *Service) GetClient(ctx context.Context, id int64) (*ClientConfig, error) {
	row := s.db.QueryRowContext(ctx, clientSelectColumns+`WHERE id = ?`, id)
	cfg, err := scanClientConfig(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query client: %w", err)
	}
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
