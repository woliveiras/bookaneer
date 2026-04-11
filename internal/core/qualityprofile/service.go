package qualityprofile

import (
	"context"
	"database/sql"
	"fmt"
)

// Service provides quality profile-related operations.
type Service struct {
	db *sql.DB
}

// New creates a new quality profile service.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// FindByID returns a quality profile by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*QualityProfile, error) {
	var qp QualityProfile
	var itemsJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, cutoff, items
		FROM quality_profiles WHERE id = ?
	`, id).Scan(&qp.ID, &qp.Name, &qp.Cutoff, &itemsJSON)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find quality profile %d: %w", id, err)
	}
	qp.Items = UnmarshalItems(itemsJSON)
	return &qp, nil
}

// List returns all quality profiles.
func (s *Service) List(ctx context.Context) ([]QualityProfile, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, cutoff, items
		FROM quality_profiles ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("list quality profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []QualityProfile
	for rows.Next() {
		var qp QualityProfile
		var itemsJSON string
		if err := rows.Scan(&qp.ID, &qp.Name, &qp.Cutoff, &itemsJSON); err != nil {
			return nil, fmt.Errorf("scan quality profile: %w", err)
		}
		qp.Items = UnmarshalItems(itemsJSON)
		profiles = append(profiles, qp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quality profiles: %w", err)
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

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO quality_profiles (name, cutoff, items)
		VALUES (?, ?, ?)
	`, input.Name, input.Cutoff, itemsJSON)
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

	_, err = s.db.ExecContext(ctx, `
		UPDATE quality_profiles SET name = ?, cutoff = ?, items = ? WHERE id = ?
	`, name, cutoff, itemsJSON, id)
	if err != nil {
		return nil, fmt.Errorf("update quality profile %d: %w", id, err)
	}

	return s.FindByID(ctx, id)
}

// Delete deletes a quality profile by ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	// Check if profile is in use by any root folder
	var inUse int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM root_folders WHERE default_quality_profile_id = ?
	`, id).Scan(&inUse)
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
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM quality_profiles").Scan(&count)
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
