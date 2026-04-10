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

func TestDetectEbookFormat(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"book.epub", "epub"},
		{"My Book EPUB Edition", "epub"},
		{"EPUB", "epub"},
		{"document.pdf", "pdf"},
		{"My Book PDF", "pdf"},
		{"book.mobi", "mobi"},
		{"MOBI edition", "mobi"},
		{"unknown.txt", "unknown"},
		{"", "unknown"},
		{"docx file", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := detectEbookFormat(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsEbookFormat(t *testing.T) {
	tests := []struct {
		give string
		want bool
	}{
		{"book.epub", true},
		{"document.pdf", true},
		{"book.mobi", true},
		{"ebook collection", true},
		{"EPUB Edition", true},
		{"PDF Download", true},
		{"MOBI Format", true},
		{"EBOOK", true},
		{"plain text", false},
		{"docx", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := isEbookFormat(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildAuthorSavePath(t *testing.T) {
	tests := []struct {
		giveDir      string
		giveAuthor   string
		giveFilename string
		wantDir      string
		wantPath     string
	}{
		{
			giveDir:      "/downloads",
			giveAuthor:   "J.R.R. Tolkien",
			giveFilename: "fellowship.epub",
			wantDir:      filepath.Join("/downloads", "J.R.R. Tolkien"),
			wantPath:     filepath.Join("/downloads", "J.R.R. Tolkien", "fellowship.epub"),
		},
		{
			giveDir:      "/downloads",
			giveAuthor:   "Author/With:Unsafe*Chars",
			giveFilename: "book.epub",
			wantDir:      filepath.Join("/downloads", "Author_With_Unsafe_Chars"),
			wantPath:     filepath.Join("/downloads", "Author_With_Unsafe_Chars", "book.epub"),
		},
		{
			giveDir:      "/books",
			giveAuthor:   "Simple Name",
			giveFilename: "my book.pdf",
			wantDir:      filepath.Join("/books", "Simple Name"),
			wantPath:     filepath.Join("/books", "Simple Name", "my book.pdf"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveAuthor, func(t *testing.T) {
			gotDir, gotPath := buildAuthorSavePath(tt.giveDir, tt.giveAuthor, tt.giveFilename)
			assert.Equal(t, tt.wantDir, gotDir)
			assert.Equal(t, tt.wantPath, gotPath)
		})
	}
}

func TestBuildAuthorSavePath_PathIsInsideAuthorDir(t *testing.T) {
	dir := t.TempDir()
	authorDir, savePath := buildAuthorSavePath(dir, "Test Author", "book.epub")
	assert.Equal(t, filepath.Join(dir, "Test Author"), authorDir)
	assert.Equal(t, filepath.Join(authorDir, "book.epub"), savePath)
}

func TestCopyFile_DestinationNotWritable(t *testing.T) {
if os.Getuid() == 0 {
t.Skip("chmod doesn't restrict root users; skipping")
}

dir := t.TempDir()

srcPath := filepath.Join(dir, "source.txt")
require.NoError(t, os.WriteFile(srcPath, []byte("content"), 0644))

// Make the destination directory read-only so os.Create fails.
destDir := filepath.Join(dir, "readonly")
require.NoError(t, os.MkdirAll(destDir, 0444))
t.Cleanup(func() { _ = os.Chmod(destDir, 0755) })

err := copyFile(srcPath, filepath.Join(destDir, "dest.txt"))
require.Error(t, err)
}
