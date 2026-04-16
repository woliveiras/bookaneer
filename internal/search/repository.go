package search

import (
	"context"
	"fmt"
	"time"
)

// ListIndexers returns all indexers.
func (s *Service) ListIndexers(ctx context.Context) ([]IndexerConfig, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, type, base_url, api_path, api_key, categories, priority, enabled,
		       enable_rss, enable_interactive_search,
		       additional_parameters, minimum_seeders, seed_ratio, seed_time,
		       created_at, updated_at
		FROM indexers ORDER BY priority ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query indexers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var indexers []IndexerConfig
	for rows.Next() {
		cfg, err := scanIndexer(rows)
		if err != nil {
			return nil, fmt.Errorf("scan indexer: %w", err)
		}
		indexers = append(indexers, cfg)
	}
	return indexers, rows.Err()
}

// GetIndexer returns an indexer by ID.
func (s *Service) GetIndexer(ctx context.Context, id int64) (*IndexerConfig, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, type, base_url, api_path, api_key, categories, priority, enabled,
		       enable_rss, enable_interactive_search,
		       additional_parameters, minimum_seeders, seed_ratio, seed_time,
		       created_at, updated_at
		FROM indexers WHERE id = ?
	`, id)
	if err != nil {
		return nil, fmt.Errorf("query indexer: %w", err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return nil, ErrIndexerNotFound
	}
	cfg, err := scanIndexer(rows)
	if err != nil {
		return nil, fmt.Errorf("scan indexer: %w", err)
	}
	return &cfg, nil
}

// CreateIndexer creates a new indexer.
func (s *Service) CreateIndexer(ctx context.Context, cfg IndexerConfig) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO indexers (name, type, base_url, api_path, api_key, categories, priority, enabled,
		    enable_rss, enable_interactive_search,
		    additional_parameters, minimum_seeders, seed_ratio, seed_time,
		    created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, cfg.Name, cfg.Type, cfg.BaseURL, cfg.APIPath, cfg.APIKey, cfg.Categories, cfg.Priority, cfg.Enabled,
		cfg.EnableRSS, cfg.EnableInteractiveSearch,
		cfg.AdditionalParameters, cfg.MinimumSeeders, cfg.SeedRatio, cfg.SeedTime,
		now, now)
	if err != nil {
		return 0, fmt.Errorf("insert indexer: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}
	cfg.ID = id
	if cfg.Enabled {
		s.loadClient(cfg)
	}
	return id, nil
}

// UpdateIndexer updates an existing indexer.
func (s *Service) UpdateIndexer(ctx context.Context, cfg IndexerConfig) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx, `
		UPDATE indexers SET name = ?, type = ?, base_url = ?, api_path = ?, api_key = ?, 
		categories = ?, priority = ?, enabled = ?,
		enable_rss = ?, enable_interactive_search = ?,
		additional_parameters = ?, minimum_seeders = ?, seed_ratio = ?, seed_time = ?,
		updated_at = ? WHERE id = ?
	`, cfg.Name, cfg.Type, cfg.BaseURL, cfg.APIPath, cfg.APIKey, cfg.Categories, cfg.Priority, cfg.Enabled,
		cfg.EnableRSS, cfg.EnableInteractiveSearch,
		cfg.AdditionalParameters, cfg.MinimumSeeders, cfg.SeedRatio, cfg.SeedTime,
		now, cfg.ID)
	if err != nil {
		return fmt.Errorf("update indexer: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrIndexerNotFound
	}
	s.mu.Lock()
	delete(s.clients, cfg.ID)
	s.mu.Unlock()
	if cfg.Enabled {
		s.loadClient(cfg)
	}
	return nil
}

// DeleteIndexer deletes an indexer.
func (s *Service) DeleteIndexer(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM indexers WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete indexer: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrIndexerNotFound
	}
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
	return nil
}

// GetOptions returns the global indexer options.
func (s *Service) GetOptions(ctx context.Context) (*IndexerOptions, error) {
	var opts IndexerOptions
	var preferFlags int
	err := s.db.QueryRowContext(ctx, `
		SELECT minimum_age, retention, maximum_size, rss_sync_interval, 
		       prefer_indexer_flags, availability_delay, updated_at
		FROM indexer_options WHERE id = 1
	`).Scan(&opts.MinimumAge, &opts.Retention, &opts.MaximumSize, &opts.RSSSyncInterval,
		&preferFlags, &opts.AvailabilityDelay, &opts.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("query indexer options: %w", err)
	}
	opts.PreferIndexerFlags = preferFlags == 1
	return &opts, nil
}

// UpdateOptions updates the global indexer options.
func (s *Service) UpdateOptions(ctx context.Context, opts IndexerOptions) error {
	now := time.Now().UTC().Format(time.RFC3339)
	preferFlags := 0
	if opts.PreferIndexerFlags {
		preferFlags = 1
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE indexer_options SET 
		    minimum_age = ?, retention = ?, maximum_size = ?, rss_sync_interval = ?,
		    prefer_indexer_flags = ?, availability_delay = ?, updated_at = ?
		WHERE id = 1
	`, opts.MinimumAge, opts.Retention, opts.MaximumSize, opts.RSSSyncInterval,
		preferFlags, opts.AvailabilityDelay, now)
	if err != nil {
		return fmt.Errorf("update indexer options: %w", err)
	}
	return nil
}
