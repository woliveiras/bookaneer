package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Role         string `json:"role"`
	APIKey       string `json:"apiKey,omitempty"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"createdAt"`
}

type Service struct {
	db *sql.DB
}

func New(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) EnsureAPIKey(ctx context.Context) error {
	var value string
	err := s.db.QueryRowContext(ctx,
		"SELECT value FROM config WHERE key = 'general.apiKey'",
	).Scan(&value)

	if err == sql.ErrNoRows || value == "" {
		apiKey, err := generateAPIKey()
		if err != nil {
			return fmt.Errorf("generate api key: %w", err)
		}
		_, err = s.db.ExecContext(ctx,
			"INSERT OR REPLACE INTO config (key, value) VALUES ('general.apiKey', ?)",
			apiKey,
		)
		if err != nil {
			return fmt.Errorf("save api key: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("check api key: %w", err)
	}
	return nil
}

func (s *Service) GetAPIKey(ctx context.Context) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		"SELECT value FROM config WHERE key = 'general.apiKey'",
	).Scan(&value)
	if err != nil {
		return "", fmt.Errorf("get api key: %w", err)
	}
	return value, nil
}

func (s *Service) ValidateAPIKey(ctx context.Context, apiKey string) bool {
	systemKey, err := s.GetAPIKey(ctx)
	if err != nil {
		return false
	}
	if apiKey == systemKey {
		return true
	}
	var id int64
	err = s.db.QueryRowContext(ctx,
		"SELECT id FROM users WHERE api_key = ?",
		apiKey,
	).Scan(&id)
	return err == nil
}

func (s *Service) CreateUser(ctx context.Context, username, password, role string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	apiKey, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("generate user api key: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO users (username, password_hash, api_key, role, created_at) VALUES (?, ?, ?, ?, ?)",
		username, string(hash), apiKey, role, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	id, _ := result.LastInsertId()
	return &User{ID: id, Username: username, Role: role, APIKey: apiKey, CreatedAt: now}, nil
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, username, password_hash, api_key, role, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.APIKey, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return &user, nil
}

func (s *Service) GetUserByAPIKey(ctx context.Context, apiKey string) (*User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		"SELECT id, username, api_key, role, created_at FROM users WHERE api_key = ?",
		apiKey,
	).Scan(&user.ID, &user.Username, &user.APIKey, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	return &user, nil
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
