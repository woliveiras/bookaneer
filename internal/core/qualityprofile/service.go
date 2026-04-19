package qualityprofile

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Service provides quality profile-related operations.
type Service struct {
	db *sqlx.DB
}

// New creates a new quality profile service.
func New(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// FindByID returns a quality profile by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*QualityProfile, error) {
	var row struct {
		ID        int64  `db:"id"`
		Name      string `db:"name"`
		Cutoff    string `db:"cutoff"`
		ItemsJSON string `db:"items"`
	}
	err := s.db.GetContext(ctx, &row, `
		SELECT id, name, cutoff, items
		FROM quality_profiles WHERE id = ?
	`, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find quality profile %d: %w", id, err)
	}
	return &QualityProfile{ID: row.ID, Name: row.Name, Cutoff: row.Cutoff, Items: UnmarshalItems(row.ItemsJSON)}, nil
}

// List returns all quality profiles.
func (s *Service) List(ctx context.Context) ([]QualityProfile, error) {
	var rows []struct {
		ID        int64  `db:"id"`
		Name      string `db:"name"`
		Cutoff    string `db:"cutoff"`
		ItemsJSON string `db:"items"`
	}
	if err := s.db.SelectContext(ctx, &rows, `
		SELECT id, name, cutoff, items
		FROM quality_profiles ORDER BY name
	`); err != nil {
		return nil, fmt.Errorf("list quality profiles: %w", err)
	}

	profiles := make([]QualityProfile, len(rows))
	for i, row := range rows {
		profiles[i] = QualityProfile{ID: row.ID, Name: row.Name, Cutoff: row.Cutoff, Items: UnmarshalItems(row.ItemsJSON)}
	}
	return profiles, nil
}

// Create creates a new quality profile.
func (s *Service) Create(ctx context.Context, input CreateQualityProfileInput) (*QualityProfile, error) {
	if input.Name == "" {
		return nil, ErrInvalidInput
	}
	if input.Cutoff == "" {
		input.Cutoff = "epub"
	}
	if len(input.Items) == 0 {
		input.Items = DefaultProfile().Items
	}

	itemsJSON := MarshalItems(input.Items)

	result, err := s.db.NamedExecContext(ctx, `
		INSERT INTO quality_profiles (name, cutoff, items)
		VALUES (:name, :cutoff, :items)
	`, map[string]any{
		"name":   input.Name,
		"cutoff": input.Cutoff,
		"items":  itemsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("create quality profile: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get quality profile id: %w", err)
	}

	return s.FindByID(ctx, id)
}

// Update updates an existing quality profile.
func (s *Service) Update(ctx context.Context, id int64, input UpdateQualityProfileInput) (*QualityProfile, error) {
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	cutoff := existing.Cutoff
	items := existing.Items

	if input.Name != nil {
		name = *input.Name
	}
	if input.Cutoff != nil {
		cutoff = *input.Cutoff
	}
	if input.Items != nil {
		items = *input.Items
	}

	itemsJSON := MarshalItems(items)

	_, err = s.db.NamedExecContext(ctx, `
		UPDATE quality_profiles SET name = :name, cutoff = :cutoff, items = :items WHERE id = :id
	`, map[string]any{
		"name":   name,
		"cutoff": cutoff,
		"items":  itemsJSON,
		"id":     id,
	})
	if err != nil {
		return nil, fmt.Errorf("update quality profile %d: %w", id, err)
	}

	return s.FindByID(ctx, id)
}

// Delete deletes a quality profile by ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	// Check if profile is in use by any root folder
	var inUse int
	err := s.db.GetContext(ctx, &inUse, `
		SELECT COUNT(*) FROM root_folders WHERE default_quality_profile_id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("check profile usage: %w", err)
	}
	if inUse > 0 {
		return ErrInUse
	}

	result, err := s.db.ExecContext(ctx, "DELETE FROM quality_profiles WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete quality profile %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// EnsureDefault ensures at least one quality profile exists.
func (s *Service) EnsureDefault(ctx context.Context) error {
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM quality_profiles")
	if err != nil {
		return fmt.Errorf("count profiles: %w", err)
	}

	if count == 0 {
		defaultProfile := DefaultProfile()
		_, err := s.Create(ctx, CreateQualityProfileInput{
			Name:   defaultProfile.Name,
			Cutoff: defaultProfile.Cutoff,
			Items:  defaultProfile.Items,
		})
		if err != nil {
			return fmt.Errorf("create default profile: %w", err)
		}
	}

	return nil
}
