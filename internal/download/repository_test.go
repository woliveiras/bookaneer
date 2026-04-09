package download_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestCreateClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	cfg := &download.ClientConfig{
		Name:    "Test SABnzbd",
		Type:    "sabnzbd",
		Host:    "localhost",
		Port:    8080,
		APIKey:  "test-key",
		Enabled: true,
	}

	err := svc.CreateClient(ctx, cfg)
	require.NoError(t, err)
	assert.NotZero(t, cfg.ID)
	assert.NotEmpty(t, cfg.CreatedAt)
}

func TestListClients_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	clients, err := svc.ListClients(ctx)
	require.NoError(t, err)
	assert.Empty(t, clients)
}

func TestListClients_WithClients(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	_ = svc.CreateClient(ctx, &download.ClientConfig{
		Name: "Client A", Type: "direct", Host: "localhost", Port: 80, Enabled: true, Priority: 2,
	})
	_ = svc.CreateClient(ctx, &download.ClientConfig{
		Name: "Client B", Type: "sabnzbd", Host: "localhost", Port: 8080, Enabled: true, Priority: 1,
	})

	clients, err := svc.ListClients(ctx)
	require.NoError(t, err)
	require.Len(t, clients, 2)
	assert.Equal(t, "Client B", clients[0].Name)
	assert.Equal(t, "Client A", clients[1].Name)
}

func TestGetClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	cfg := &download.ClientConfig{
		Name:     "Test Client",
		Type:     "direct",
		Host:     "example.com",
		Port:     443,
		UseTLS:   true,
		Username: "user",
		Password: "pass",
		Enabled:  true,
	}
	err := svc.CreateClient(ctx, cfg)
	require.NoError(t, err)

	got, err := svc.GetClient(ctx, cfg.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test Client", got.Name)
	assert.Equal(t, "direct", got.Type)
	assert.Equal(t, "example.com", got.Host)
	assert.Equal(t, "user", got.Username)
	assert.Equal(t, "pass", got.Password)
}

func TestGetClient_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	_, err := svc.GetClient(ctx, 9999)
	require.Error(t, err)
}

func TestUpdateClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	cfg := &download.ClientConfig{
		Name: "Original", Type: "direct", Host: "localhost", Port: 80, Enabled: true,
	}
	err := svc.CreateClient(ctx, cfg)
	require.NoError(t, err)

	cfg.Name = "Updated"
	cfg.Host = "newhost.com"
	err = svc.UpdateClient(ctx, cfg)
	require.NoError(t, err)

	got, err := svc.GetClient(ctx, cfg.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", got.Name)
	assert.Equal(t, "newhost.com", got.Host)
}

func TestDeleteClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	cfg := &download.ClientConfig{
		Name: "To Delete", Type: "direct", Host: "localhost", Port: 80, Enabled: true,
	}
	err := svc.CreateClient(ctx, cfg)
	require.NoError(t, err)

	err = svc.DeleteClient(ctx, cfg.ID)
	require.NoError(t, err)

	_, err = svc.GetClient(ctx, cfg.ID)
	require.Error(t, err)
}

func TestDeleteClient_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	err := svc.DeleteClient(ctx, 9999)
	require.Error(t, err)
}

func TestCreateClient_NullableFields(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	cfg := &download.ClientConfig{
		Name:    "Minimal",
		Type:    "direct",
		Host:    "localhost",
		Port:    80,
		Enabled: true,
	}
	err := svc.CreateClient(ctx, cfg)
	require.NoError(t, err)

	got, err := svc.GetClient(ctx, cfg.ID)
	require.NoError(t, err)
	assert.Equal(t, "", got.Username)
	assert.Equal(t, "", got.Password)
	assert.Equal(t, "", got.APIKey)
	assert.Equal(t, "", got.DownloadDir)
}
