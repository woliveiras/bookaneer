package rootfolder_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/rootfolder"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestMoveRootFolder_SamePath(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)
	ctx := context.Background()

	dir := t.TempDir()
	rf, err := svc.Create(ctx, rootfolder.CreateRootFolderInput{Path: dir, Name: "Library"})
	require.NoError(t, err)

	result, err := svc.MoveRootFolder(ctx, rf.ID, dir)
	require.NoError(t, err)
	assert.Equal(t, dir, result.Path)
}

func TestMoveRootFolder_WithFiles(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	newDir := filepath.Join(t.TempDir(), "new_library")

	authorDir := filepath.Join(oldDir, "Tolkien")
	require.NoError(t, os.MkdirAll(authorDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(authorDir, "hobbit.epub"), []byte("ebook data"), 0644))

	rf, err := svc.Create(ctx, rootfolder.CreateRootFolderInput{Path: oldDir, Name: "Library"})
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO authors (name, sort_name, monitored, path) VALUES ('Tolkien', 'Tolkien', 1, ?)", authorDir)
	require.NoError(t, err)

	result, err := svc.MoveRootFolder(ctx, rf.ID, newDir)
	require.NoError(t, err)
	assert.Equal(t, newDir, result.Path)

	_, err = os.Stat(filepath.Join(newDir, "Tolkien", "hobbit.epub"))
	require.NoError(t, err)
}

func TestMoveRootFolder_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)

	_, err := svc.MoveRootFolder(context.Background(), 9999, "/tmp/new")
	require.ErrorIs(t, err, rootfolder.ErrNotFound)
}

func TestUpdate_WithPath(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	rf, err := svc.Create(ctx, rootfolder.CreateRootFolderInput{Path: oldDir, Name: "Old"})
	require.NoError(t, err)

	newDir := t.TempDir()
	updated, err := svc.Update(ctx, rf.ID, rootfolder.UpdateRootFolderInput{Path: &newDir})
	require.NoError(t, err)
	assert.Equal(t, newDir, updated.Path)
}

func TestUpdate_MoveFiles(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)
	ctx := context.Background()

	oldDir := t.TempDir()
	newDir := filepath.Join(t.TempDir(), "moved")

	rf, err := svc.Create(ctx, rootfolder.CreateRootFolderInput{Path: oldDir, Name: "Library"})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, rf.ID, rootfolder.UpdateRootFolderInput{
		Path:      &newDir,
		MoveFiles: true,
	})
	require.NoError(t, err)
	assert.Equal(t, newDir, updated.Path)
}

func TestUpdate_NameOnly(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)
	ctx := context.Background()

	dir := t.TempDir()
	rf, err := svc.Create(ctx, rootfolder.CreateRootFolderInput{Path: dir, Name: "Old"})
	require.NoError(t, err)

	newName := "New Name"
	updated, err := svc.Update(ctx, rf.ID, rootfolder.UpdateRootFolderInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestUpdate_NoChanges(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)
	ctx := context.Background()

	dir := t.TempDir()
	rf, err := svc.Create(ctx, rootfolder.CreateRootFolderInput{Path: dir, Name: "Library"})
	require.NoError(t, err)

	result, err := svc.Update(ctx, rf.ID, rootfolder.UpdateRootFolderInput{})
	require.NoError(t, err)
	assert.Equal(t, rf.ID, result.ID)
}

func TestCreate_PathNotExist_CreatesDir(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := rootfolder.New(db)
	ctx := context.Background()

	dir := filepath.Join(t.TempDir(), "nonexistent", "library")
	rf, err := svc.Create(ctx, rootfolder.CreateRootFolderInput{Path: dir, Name: "Auto Created"})
	require.NoError(t, err)
	assert.NotZero(t, rf.ID)

	_, err = os.Stat(dir)
	require.NoError(t, err)
}
