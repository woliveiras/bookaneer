package wanted

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal Name", "Normal Name"},
		{"Name/With/Slashes", "Name_With_Slashes"},
		{"Name\\Backslash", "Name_Backslash"},
		{"Name:Colon", "Name_Colon"},
		{"Name*Star", "Name_Star"},
		{"Name?Question", "Name_Question"},
		{"Name\"Quote", "Name_Quote"},
		{"Name<Less>Greater", "Name_Less_Greater"},
		{"Name|Pipe", "Name_Pipe"},
		{"All/\\:*?\"<>|Chars", "All_________Chars"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.txt")
	dstPath := filepath.Join(dir, "dest.txt")

	content := []byte("hello world")
	require.NoError(t, os.WriteFile(srcPath, content, 0644))

	err := copyFile(srcPath, dstPath)
	require.NoError(t, err)

	got, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestCopyFile_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	err := copyFile(filepath.Join(dir, "missing"), filepath.Join(dir, "dest"))
	require.Error(t, err)
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0644))

	hash := hashFile(path)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64) // SHA256 hex = 64 chars
	// Known SHA256 of "hello"
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hash)
}

func TestHashFile_MissingFile(t *testing.T) {
	hash := hashFile("/nonexistent/file")
	assert.Empty(t, hash)
}
