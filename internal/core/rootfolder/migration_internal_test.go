package rootfolder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

// ── copyFile ─────────────────────────────────────────────────────────────────

func TestCopyFile_Success(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "src.epub")
	dst := filepath.Join(t.TempDir(), "dst.epub")
	content := []byte("ebook content")
	require.NoError(t, os.WriteFile(src, content, 0644))

	require.NoError(t, copyFile(src, dst))

	got, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestCopyFile_PreservesPermissions(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "src.epub")
	dst := filepath.Join(t.TempDir(), "dst.epub")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0600))

	require.NoError(t, copyFile(src, dst))

	info, err := os.Stat(dst)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestCopyFile_SourceNotExist(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "nonexistent.epub")
	dst := filepath.Join(t.TempDir(), "dst.epub")

	err := copyFile(src, dst)
	require.Error(t, err)
}

func TestCopyFile_DestDirNotExist(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "src.epub")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0644))

	dst := filepath.Join(t.TempDir(), "missing_dir", "dst.epub")

	err := copyFile(src, dst)
	require.Error(t, err)
}

// ── copyDir ──────────────────────────────────────────────────────────────────

func TestCopyDir_WithFiles(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")

	require.NoError(t, os.WriteFile(filepath.Join(src, "book1.epub"), []byte("ebook1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(src, "book2.epub"), []byte("ebook2"), 0644))

	require.NoError(t, copyDir(src, dst))

	got1, err := os.ReadFile(filepath.Join(dst, "book1.epub"))
	require.NoError(t, err)
	assert.Equal(t, []byte("ebook1"), got1)

	got2, err := os.ReadFile(filepath.Join(dst, "book2.epub"))
	require.NoError(t, err)
	assert.Equal(t, []byte("ebook2"), got2)
}

func TestCopyDir_Nested(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")

	subdir := filepath.Join(src, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "nested.epub"), []byte("nested"), 0644))

	require.NoError(t, copyDir(src, dst))

	got, err := os.ReadFile(filepath.Join(dst, "subdir", "nested.epub"))
	require.NoError(t, err)
	assert.Equal(t, []byte("nested"), got)
}

func TestCopyDir_Empty(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")

	require.NoError(t, copyDir(src, dst))

	info, err := os.Stat(dst)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestCopyDir_SourceNotExist(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "nonexistent")
	dst := filepath.Join(t.TempDir(), "dst")

	err := copyDir(src, dst)
	require.Error(t, err)
}

// TestCopyDir_DstMkdirFails covers the MkdirAll(dst) error branch: src exists
// but the destination can't be created (read-only parent directory).
func TestCopyDir_DstMkdirFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can write anywhere; skipping")
	}

	src := t.TempDir()
	base := t.TempDir()
	readOnly := filepath.Join(base, "readonly")
	require.NoError(t, os.Mkdir(readOnly, 0555))
	t.Cleanup(func() { _ = os.Chmod(readOnly, 0755) })

	dst := filepath.Join(readOnly, "dst")
	err := copyDir(src, dst)
	require.Error(t, err)
}

// ── MoveRootFolder (additional paths) ────────────────────────────────────────

// TestMoveRootFolder_AuthorPathMissing covers the warn-and-continue branch
// inside MoveRootFolder when an author's directory does not exist on disk.
func TestMoveRootFolder_AuthorPathMissing(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	newDir := filepath.Join(t.TempDir(), "new_library")

	rf, err := svc.Create(ctx, CreateRootFolderInput{Path: oldDir, Name: "Library"})
	require.NoError(t, err)

	// Insert an author whose path does not exist on disk.
	ghostPath := filepath.Join(oldDir, "Ghost_Author")
	_, err = db.ExecContext(ctx,
		"INSERT INTO authors (name, sort_name, monitored, path) VALUES ('Ghost', 'Ghost', 1, ?)",
		ghostPath,
	)
	require.NoError(t, err)

	// MoveRootFolder should succeed even when the author directory is absent.
	result, err := svc.MoveRootFolder(ctx, rf.ID, newDir)
	require.NoError(t, err)
	assert.Equal(t, newDir, result.Path)

	// The author's path in the DB should still be updated.
	var updatedPath string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT path FROM authors WHERE name = 'Ghost'").Scan(&updatedPath))
	assert.Equal(t, filepath.Join(newDir, "Ghost_Author"), updatedPath)
}

// TestMoveRootFolder_UpdatesBookFilePaths verifies that book_files rows with
// paths inside the old root folder are rewritten to the new location.
func TestMoveRootFolder_UpdatesBookFilePaths(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	newDir := filepath.Join(t.TempDir(), "new_library")

	authorDir := filepath.Join(oldDir, "Tolkien")
	require.NoError(t, os.MkdirAll(authorDir, 0755))

	rf, err := svc.Create(ctx, CreateRootFolderInput{Path: oldDir, Name: "Library"})
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		"INSERT INTO authors (name, sort_name, monitored, path) VALUES ('Tolkien', 'Tolkien', 1, ?)",
		authorDir,
	)
	require.NoError(t, err)

	var authorID int64
	require.NoError(t, db.QueryRowContext(ctx, "SELECT id FROM authors WHERE name = 'Tolkien'").Scan(&authorID))

	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")
	bookFilePath := filepath.Join(authorDir, "hobbit.epub")
	require.NoError(t, os.WriteFile(bookFilePath, []byte("epub"), 0644))

	_, err = db.ExecContext(ctx,
		"INSERT INTO book_files (book_id, path, relative_path, format, size) VALUES (?, ?, ?, 'epub', 100)",
		bookID, bookFilePath, "Tolkien/hobbit.epub",
	)
	require.NoError(t, err)

	result, err := svc.MoveRootFolder(ctx, rf.ID, newDir)
	require.NoError(t, err)
	assert.Equal(t, newDir, result.Path)

	var updatedFilePath string
	require.NoError(t, db.QueryRowContext(ctx, "SELECT path FROM book_files WHERE book_id = ?", bookID).Scan(&updatedFilePath))
	assert.Equal(t, filepath.Join(newDir, "Tolkien", "hobbit.epub"), updatedFilePath)
}

// ── copyFile error paths ──────────────────────────────────────────────────────

// TestCopyFile_Unreadable covers the os.Open error path inside copyFile when
// the source is not readable.
func TestCopyFile_Unreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can read mode-000 files; skipping")
	}
	t.Parallel()

	src := filepath.Join(t.TempDir(), "secret.epub")
	dst := filepath.Join(t.TempDir(), "dst.epub")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0000))

	err := copyFile(src, dst)
	require.Error(t, err)
}

// ── copyDir error paths ───────────────────────────────────────────────────────

// TestCopyDir_FileUnreadable covers the branch where copyFile fails because
// a file inside src is unreadable (mode 000).
func TestCopyDir_FileUnreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can read mode-000 files; skipping")
	}

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")
	require.NoError(t, os.WriteFile(filepath.Join(src, "secret.epub"), []byte("data"), 0000))

	err := copyDir(src, dst)
	require.Error(t, err)
}

// TestCopyDir_NestedFileUnreadable covers the branch where the recursive
// copyDir call fails due to an unreadable file inside a subdirectory.
func TestCopyDir_NestedFileUnreadable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can read mode-000 files; skipping")
	}

	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dst")

	subdir := filepath.Join(src, "subdir")
	require.NoError(t, os.MkdirAll(subdir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subdir, "secret.epub"), []byte("data"), 0000))

	err := copyDir(src, dst)
	require.Error(t, err)
}

// TestMoveRootFolder_NewDirCreateFails covers the branch where os.MkdirAll
// fails when creating the target directory for the migration.
func TestMoveRootFolder_NewDirCreateFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can write anywhere; skipping")
	}

	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	rf, err := svc.Create(ctx, CreateRootFolderInput{Path: oldDir, Name: "Library"})
	require.NoError(t, err)

	base := t.TempDir()
	readOnly := filepath.Join(base, "readonly")
	require.NoError(t, os.Mkdir(readOnly, 0555))
	t.Cleanup(func() { _ = os.Chmod(readOnly, 0755) })

	newPath := filepath.Join(readOnly, "newlibrary")
	_, err = svc.MoveRootFolder(ctx, rf.ID, newPath)
	require.Error(t, err)
}

// TestMoveRootFolder_DBWritesFail exercises the transaction error paths in
// MoveRootFolder by setting the SQLite connection to query-only mode after
// the read-phase succeeds.
func TestMoveRootFolder_DBWritesFail(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	newDir := filepath.Join(t.TempDir(), "new_library")

	rf, err := svc.Create(ctx, CreateRootFolderInput{Path: oldDir, Name: "Library"})
	require.NoError(t, err)

	// Disable all writes on this connection.  MoveRootFolder will read the
	// existing row and filesystem successfully, then fail when it tries to
	// BEGIN / commit the UPDATE transaction.
	_, err = db.ExecContext(ctx, "PRAGMA query_only = ON")
	require.NoError(t, err)

	_, err = svc.MoveRootFolder(ctx, rf.ID, newDir)
	require.Error(t, err)
}

// ── Service DB-error paths ────────────────────────────────────────────────────

func TestList_DBError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	_ = db.Close()
	_, err := s.List(context.Background())
	require.Error(t, err)
}

func TestFindByID_DBError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	_ = db.Close()
	_, err := s.FindByID(context.Background(), 1)
	require.Error(t, err)
}

func TestCreate_DBError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	dir := t.TempDir() // path exists, so os.Stat succeeds
	_ = db.Close()
	_, err := s.Create(context.Background(), CreateRootFolderInput{Path: dir, Name: "Test"})
	require.Error(t, err)
}

func TestDelete_DBError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	_ = db.Close()
	err := s.Delete(context.Background(), 1)
	require.Error(t, err)
}

// TestUpdate_DBWritesFail covers the general "update root folder" error branch
// by switching the connection to query-only mode after the initial read.
func TestUpdate_DBWritesFail(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "Original"})
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "PRAGMA query_only = ON")
	require.NoError(t, err)

	newName := "Updated"
	_, err = s.Update(ctx, created.ID, UpdateRootFolderInput{Name: &newName})
	require.Error(t, err)
}

// TestMoveRootFolder_QueryAuthorsFails covers the "query authors" error branch
// by renaming the authors table so the SELECT inside MoveRootFolder fails.
func TestMoveRootFolder_QueryAuthorsFails(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	newDir := filepath.Join(t.TempDir(), "new_library")

	rf, err := svc.Create(ctx, CreateRootFolderInput{Path: oldDir, Name: "Library"})
	require.NoError(t, err)

	// Disable FK checks then rename to simulate a missing table without
	// violating the foreign-key that books have on authors.
	_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = OFF")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "ALTER TABLE authors RENAME TO authors_bak")
	require.NoError(t, err)

	_, err = svc.MoveRootFolder(ctx, rf.ID, newDir)
	require.Error(t, err)
}

// TestCopyDir_ReadDirFails covers the ReadDir error path: src exists (stat OK)
// but directory listing is denied (no read bit on the directory).
func TestCopyDir_ReadDirFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can read everything; skipping")
	}

	src := t.TempDir()
	// mode 0100 (execute only, no read): os.Stat still succeeds (uses parent's
	// execute bit), but os.ReadDir fails because it needs the read bit.
	require.NoError(t, os.Chmod(src, 0100))
	t.Cleanup(func() { _ = os.Chmod(src, 0755) }) // restore so t.TempDir cleanup works

	dst := filepath.Join(t.TempDir(), "dst")
	err := copyDir(src, dst)
	require.Error(t, err)
}
