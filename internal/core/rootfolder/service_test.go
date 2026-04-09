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
	db := testutil.OpenTestDB(t)
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
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.Create(ctx, CreateRootFolderInput{Path: "", Name: ""})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestCreate_DuplicatePath(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	dir := t.TempDir()
	_, err := s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "A"})
	require.NoError(t, err)

	_, err = s.Create(ctx, CreateRootFolderInput{Path: dir, Name: "B"})
	assert.ErrorIs(t, err, ErrDuplicate)
}

func TestCreate_CreatesDirectory(t *testing.T) {
	db := testutil.OpenTestDB(t)
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
	db := testutil.OpenTestDB(t)
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
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.FindByID(ctx, 999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList(t *testing.T) {
	db := testutil.OpenTestDB(t)
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
	db := testutil.OpenTestDB(t)
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
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	name := "Ghost"
	_, err := s.Update(ctx, 999, UpdateRootFolderInput{Name: &name})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete(t *testing.T) {
	db := testutil.OpenTestDB(t)
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
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	err := s.Delete(ctx, 999)
	assert.ErrorIs(t, err, ErrNotFound)
}
