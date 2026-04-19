package author

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestSanitizeFolderName(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"Normal Name", "Normal Name"},
		{"Name/Slash", "Name-Slash"},
		{"Name\\Backslash", "Name-Backslash"},
		{"Name:Colon", "Name-Colon"},
		{"Name*Star", "NameStar"},
		{"Name?Question", "NameQuestion"},
		{`Name"Quote`, "Name'Quote"},
		{"Name<Less", "NameLess"},
		{"Name>Greater", "NameGreater"},
		{"Name|Pipe", "Name-Pipe"},
		{"  Trimmed  ", "Trimmed"},
		{"All/\\:*?\"<>|Chars", "All---'-Chars"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := sanitizeFolderName(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteAuthorFiles_NoRootFolder(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := New(db)
	ctx := context.Background()

	// No root_folders row — should return nil (nothing to delete)
	a := &Author{ID: 1, Name: "Ghost Author"}
	err := svc.deleteAuthorFiles(ctx, a)
	require.NoError(t, err)
}

func TestDeleteAuthorFiles_FolderDoesNotExist(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	testutil.SeedRootFolder(t, db.DB, dir, "Library")

	// Author folder does not exist on disk — should return nil
	a := &Author{ID: 1, Name: "Nonexistent Author"}
	err := svc.deleteAuthorFiles(ctx, a)
	require.NoError(t, err)
}

func TestDeleteAuthorFiles_RemovesFolder(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	testutil.SeedRootFolder(t, db.DB, dir, "Library")

	// Create the author folder with a file inside
	authorFolderName := sanitizeFolderName("Brandon Sanderson")
	authorPath := filepath.Join(dir, authorFolderName)
	require.NoError(t, os.MkdirAll(authorPath, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(authorPath, "book.epub"), []byte("data"), 0o644))

	a := &Author{ID: 1, Name: "Brandon Sanderson"}
	err := svc.deleteAuthorFiles(ctx, a)
	require.NoError(t, err)

	_, statErr := os.Stat(authorPath)
	assert.True(t, os.IsNotExist(statErr), "author folder should have been removed")
}

func TestDeleteAuthorFiles_DBError(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := New(db)
	ctx := context.Background()

	// Close the DB so the root_folder query returns a non-ErrNoRows error,
	// which surfaces the "get root folder" error branch.
	_ = db.Close()

	a := &Author{ID: 1, Name: "Ghost"}
	err := svc.deleteAuthorFiles(ctx, a)
	require.Error(t, err)
}

func TestDeleteAuthorFiles_RemoveAllError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test RemoveAll errors as root")
	}

	db := testutil.OpenTestDBX(t)
	svc := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	testutil.SeedRootFolder(t, db.DB, dir, "Library")

	authorFolderName := sanitizeFolderName("Protected Author")
	authorPath := filepath.Join(dir, authorFolderName)
	require.NoError(t, os.MkdirAll(authorPath, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(authorPath, "book.epub"), []byte("data"), 0o644))

	// Remove write permission from parent dir so RemoveAll cannot unlink authorPath.
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	a := &Author{ID: 1, Name: "Protected Author"}
	err := svc.deleteAuthorFiles(ctx, a)
	require.Error(t, err)
}

func TestGetStats_TotalSize(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Isaac Asimov")
	bookID1 := testutil.SeedBook(t, db.DB, authorID, "Foundation")
	bookID2 := testutil.SeedBook(t, db.DB, authorID, "I Robot")

	_, err := db.ExecContext(ctx,
		`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, ?, ?, ?, 'epub', 'epub')`,
		bookID1, "/lib/foundation.epub", "foundation.epub", 3000)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, ?, ?, ?, 'epub', 'epub')`,
		bookID2, "/lib/irobot.epub", "irobot.epub", 2000)
	require.NoError(t, err)

	stats, err := svc.GetStats(ctx, authorID)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.BookCount)
	assert.Equal(t, 2, stats.BookFileCount)
	assert.Equal(t, 0, stats.MissingBooks) // both books have files
	assert.Equal(t, 5000, stats.TotalSizeBytes)
}
