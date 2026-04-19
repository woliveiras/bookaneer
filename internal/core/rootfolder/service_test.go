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

func TestCreate(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	rf, err := s.Create(ctx, CreateRootFolderInput{
		Path: dir,
		Name: "Library",
	})
	require.NoError(t, err)
	assert.Greater(t, rf.ID, int64(0))
	assert.Equal(t, "Library", rf.Name)
	assert.Equal(t, dir, rf.Path)
}

func TestCreate_MissingFields(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.Create(ctx, CreateRootFolderInput{Path: "", Name: ""})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestCreate_DuplicatePath(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	_, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "A"})
	require.NoError(t, err)

	_, err = s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "B"})
	assert.ErrorIs(t, err, ErrDuplicate)
}

func TestCreate_CreatesDirectory(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := filepath.Join(t.TempDir(), "newdir")
	_, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "New"})
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFindByID(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "Test"})
	require.NoError(t, err)

	got, err := s.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test", got.Name)
	assert.True(t, got.Accessible)
}

func TestFindByID_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.FindByID(ctx, 999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.Create(ctx, CreateRootFolderInput{
		Path: filepath.Join(t.TempDir(), "a"),
		Name: "Alpha",
	})
	require.NoError(t, err)
	_, err = s.Create(ctx, CreateRootFolderInput{
		Path: filepath.Join(t.TempDir(), "b"),
		Name: "Beta",
	})
	require.NoError(t, err)

	folders, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, folders, 2)
	assert.Equal(t, "Alpha", folders[0].Name)
}

func TestUpdate(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "Old"})
	require.NoError(t, err)

	newName := "New"
	updated, err := s.Update(ctx, created.ID, UpdateRootFolderInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
}

func TestUpdate_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	name := "Ghost"
	_, err := s.Update(ctx, 999, UpdateRootFolderInput{Name: &name})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "ToDelete"})
	require.NoError(t, err)

	err = s.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = s.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	err := s.Delete(ctx, 999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestCreate_PathIsFile(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	// Create a file (not a directory)
	filePath := filepath.Join(t.TempDir(), "notadir")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o644))

	_, err := s.Create(ctx, CreateRootFolderInput{Path: filePath, Name: "Bad"})
	require.ErrorIs(t, err, ErrPathNotAccessible)
}

func TestUpdate_PathUpdate(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "Original"})
	require.NoError(t, err)

	newDir := t.TempDir()
	updated, err := s.Update(ctx, created.ID, UpdateRootFolderInput{Path: &newDir})
	require.NoError(t, err)
	assert.Equal(t, newDir, updated.Path)
}

func TestUpdate_PathIsFile(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "Test"})
	require.NoError(t, err)

	filePath := filepath.Join(t.TempDir(), "notadir")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o644))

	_, err = s.Update(ctx, created.ID, UpdateRootFolderInput{Path: &filePath})
	require.ErrorIs(t, err, ErrPathNotAccessible)
}

func TestUpdate_DefaultQualityProfileID(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	// Create a quality profile first (FK constraint)
	_, err := db.ExecContext(ctx, `INSERT INTO quality_profiles (name, cutoff, items) VALUES ('Test QP', 'epub', '[]')`)
	require.NoError(t, err)
	var qpID int64
	require.NoError(t, db.QueryRowContext(ctx, "SELECT id FROM quality_profiles LIMIT 1").Scan(&qpID))

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "QP"})
	require.NoError(t, err)

	updated, err := s.Update(ctx, created.ID, UpdateRootFolderInput{DefaultQualityProfileID: &qpID})
	require.NoError(t, err)
	require.NotNil(t, updated.DefaultQualityProfileID)
	assert.Equal(t, qpID, *updated.DefaultQualityProfileID)
}

func TestUpdate_EmptyNoChanges(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "NoOp"})
	require.NoError(t, err)

	updated, err := s.Update(ctx, created.ID, UpdateRootFolderInput{})
	require.NoError(t, err)
	assert.Equal(t, created.Name, updated.Name)
}

func TestUpdate_DuplicatePath(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	_, err := s.Create(ctx, CreateRootFolderInput{Path: dir1, Name: "First"})
	require.NoError(t, err)
	second, err := s.Create(ctx, CreateRootFolderInput{Path: dir2, Name: "Second"})
	require.NoError(t, err)

	_, err = s.Update(ctx, second.ID, UpdateRootFolderInput{Path: &dir1})
	require.ErrorIs(t, err, ErrDuplicate)
}

func TestEnrichWithDiskInfo_Inaccessible(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)

	rf := &RootFolder{Path: "/nonexistent/path/that/does/not/exist"}
	s.enrichWithDiskInfo(rf)
	assert.False(t, rf.Accessible)
}

func TestEnrichWithDiskInfo_Accessible(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)

	rf := &RootFolder{Path: t.TempDir()}
	s.enrichWithDiskInfo(rf)
	assert.True(t, rf.Accessible)
	assert.Greater(t, rf.TotalSpace, int64(0))
}

// ── Permission / filesystem error paths ──────────────────────────────────────

// TestCreate_StatError covers the branch where os.Stat returns an error that
// is not os.IsNotExist (e.g., ENOTDIR when a path component is a file).
func TestCreate_StatError(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	// Make a regular file, then use it as a path prefix so os.Stat on the
	// child path gets ENOTDIR — not a "not found" error.
	fileAsDir := filepath.Join(t.TempDir(), "notadir")
	require.NoError(t, os.WriteFile(fileAsDir, []byte("data"), 0o644))

	path := filepath.Join(fileAsDir, "subdir")
	_, err := s.Create(ctx, CreateRootFolderInput{Path: path, Name: "Test"})
	require.ErrorIs(t, err, ErrPathNotAccessible)
}

// TestCreate_MkdirAllFails covers the branch where the path does not exist but
// os.MkdirAll fails due to a permissions problem on the parent directory.
func TestCreate_MkdirAllFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can write anywhere; skipping")
	}

	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	base := t.TempDir()
	readOnly := filepath.Join(base, "readonly")
	require.NoError(t, os.Mkdir(readOnly, 0o555))
	t.Cleanup(func() { _ = os.Chmod(readOnly, 0o755) }) // restore for cleanup

	path := filepath.Join(readOnly, "newlib")
	_, err := s.Create(ctx, CreateRootFolderInput{Path: path, Name: "Test"})
	require.Error(t, err)
}

// TestUpdate_MkdirAllFails covers the branch in Update where the new path
// doesn't exist and MkdirAll fails (read-only parent directory).
func TestUpdate_MkdirAllFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can write anywhere; skipping")
	}

	db := testutil.OpenTestDBX(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "Original"})
	require.NoError(t, err)

	base := t.TempDir()
	readOnly := filepath.Join(base, "readonly")
	require.NoError(t, os.Mkdir(readOnly, 0o555))
	t.Cleanup(func() { _ = os.Chmod(readOnly, 0o755) })

	newPath := filepath.Join(readOnly, "newlib")
	_, err = s.Update(ctx, created.ID, UpdateRootFolderInput{Path: &newPath})
	require.Error(t, err)
}
