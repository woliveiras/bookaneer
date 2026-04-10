package auth

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePassword_Length(t *testing.T) {
	t.Parallel()

	pwd, err := generatePassword(16)
	require.NoError(t, err)
	assert.Len(t, pwd, 16)
}

func TestGeneratePassword_LongLength(t *testing.T) {
	t.Parallel()

	pwd, err := generatePassword(64)
	require.NoError(t, err)
	assert.Len(t, pwd, 64)
}

func TestGeneratePassword_OnlyValidChars(t *testing.T) {
	t.Parallel()

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	pwd, err := generatePassword(256)
	require.NoError(t, err)
	for _, c := range pwd {
		assert.Contains(t, charset, string(c), "unexpected character %q in password", c)
	}
}

func TestGenerateAPIKey_Length(t *testing.T) {
	t.Parallel()

	key, err := generateAPIKey()
	require.NoError(t, err)
	// 32 random bytes → 64 hex characters
	assert.Len(t, key, 64)
}

func TestGenerateAPIKey_ValidHex(t *testing.T) {
	t.Parallel()

	key, err := generateAPIKey()
	require.NoError(t, err)
	_, decodeErr := hex.DecodeString(key)
	assert.NoError(t, decodeErr, "API key should be valid hex")
}

func TestGenerateAPIKey_Unique(t *testing.T) {
	t.Parallel()

	key1, err := generateAPIKey()
	require.NoError(t, err)
	key2, err := generateAPIKey()
	require.NoError(t, err)
	assert.NotEqual(t, key1, key2)
}
