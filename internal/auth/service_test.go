package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/auth"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestCreateUser(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "testuser", "password123", "admin")
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "admin", user.Role)
	assert.NotEmpty(t, user.APIKey)
	assert.Len(t, user.APIKey, 64)
	assert.NotEmpty(t, user.CreatedAt)
}

func TestAuthenticate_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "admin", "secret", "admin")
	require.NoError(t, err)

	user, err := svc.Authenticate(ctx, "admin", "secret")
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.NotEmpty(t, user.APIKey)
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "admin", "secret", "admin")
	require.NoError(t, err)

	_, err = svc.Authenticate(ctx, "admin", "wrongpass")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	_, err := svc.Authenticate(ctx, "nobody", "pass")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestGetUserByAPIKey_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	created, err := svc.CreateUser(ctx, "admin", "pass", "admin")
	require.NoError(t, err)

	found, err := svc.GetUserByAPIKey(ctx, created.APIKey)
	require.NoError(t, err)
	assert.Equal(t, "admin", found.Username)
	assert.Equal(t, created.ID, found.ID)
}

func TestGetUserByAPIKey_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	_, err := svc.GetUserByAPIKey(ctx, "nonexistent_key")
	require.ErrorIs(t, err, auth.ErrUserNotFound)
}

func TestEnsureAPIKey(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	err := svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	key, err := svc.GetAPIKey(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.Len(t, key, 64)

	err = svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	key2, err := svc.GetAPIKey(ctx)
	require.NoError(t, err)
	assert.Equal(t, key, key2)
}

func TestValidateAPIKey_SystemKey(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	err := svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	key, err := svc.GetAPIKey(ctx)
	require.NoError(t, err)

	assert.True(t, svc.ValidateAPIKey(ctx, key))
	assert.False(t, svc.ValidateAPIKey(ctx, "invalid_key"))
}

func TestGetAPIKey_BeforeGenerate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	// Config table has empty apiKey by default from migration
	key, err := svc.GetAPIKey(context.Background())
	require.NoError(t, err)
	assert.Empty(t, key)
}

func TestValidateAPIKey_UserKey(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	err := svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	user, err := svc.CreateUser(ctx, "admin", "pass", "admin")
	require.NoError(t, err)

	assert.True(t, svc.ValidateAPIKey(ctx, user.APIKey))
}

func TestEnsureDefaultAdmin_CreatesAdmin(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	password, err := svc.EnsureDefaultAdmin(ctx, "")
	require.NoError(t, err)
	assert.NotEmpty(t, password)

	user, err := svc.Authenticate(ctx, "admin", password)
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
}

func TestEnsureDefaultAdmin_WithEnvPassword(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	password, err := svc.EnsureDefaultAdmin(ctx, "mypassword")
	require.NoError(t, err)
	assert.Equal(t, "mypassword", password)

	user, err := svc.Authenticate(ctx, "admin", "mypassword")
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
}

func TestEnsureDefaultAdmin_AlreadyExists(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "existinguser", "pass", "admin")
	require.NoError(t, err)

	password, err := svc.EnsureDefaultAdmin(ctx, "")
	require.NoError(t, err)
	assert.Empty(t, password)
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	_, err := svc.CreateUser(ctx, "dup_user", "pass1", "admin")
	require.NoError(t, err)

	_, err = svc.CreateUser(ctx, "dup_user", "pass2", "admin")
	require.Error(t, err)
}

func TestValidateAPIKey_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	// After ensuring a key, empty string should not validate
	err := svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	assert.False(t, svc.ValidateAPIKey(ctx, ""))
}

func TestValidateAPIKey_RandomInvalid(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	err := svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	assert.False(t, svc.ValidateAPIKey(ctx, "completely-invalid-key-12345"))
}

func TestGetAPIKey_AfterEnsure(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	err := svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	key, err := svc.GetAPIKey(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.Len(t, key, 64) // 32 bytes hex-encoded
}

func TestEnsureAPIKey_Idempotent(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := auth.New(db)
	ctx := context.Background()

	err := svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	key1, _ := svc.GetAPIKey(ctx)

	err = svc.EnsureAPIKey(ctx)
	require.NoError(t, err)

	key2, _ := svc.GetAPIKey(ctx)
	assert.Equal(t, key1, key2)
}
