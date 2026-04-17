package download

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// getOrCreateClient returns a cached client or creates a new one.
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

	// Inject the bypass service so direct clients can resolve challenges.
	cfg.Bypasser = s.bypasser
	cfg.CertificateValidation = s.certValidation

	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	s.clients[cfg.ID] = client
	return client, nil
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
		Name:                  "Embedded Downloader",
		Type:                  ClientTypeDirect,
		DownloadDir:           rootPath,
		Enabled:               true,
		Bypasser:              s.bypasser,
		CertificateValidation: s.certValidation,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create embedded client: %w", err)
	}

	s.embeddedClient = client
	s.embeddedClientConfig = &ClientConfig{
		ID:                    0,
		Name:                  "Embedded Downloader",
		Type:                  ClientTypeDirect,
		DownloadDir:           rootPath,
		Enabled:               true,
		Bypasser:              s.bypasser,
		CertificateValidation: s.certValidation,
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

	// placeholders contains only "?" — not user input — so fmt.Sprintf is safe here.
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
	defer func() { _ = rows.Close() }()

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
