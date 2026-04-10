package author_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/author"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestFindByID_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	db.Close()

	_, err := svc.FindByID(ctx, 1)
	require.Error(t, err)
	require.NotErrorIs(t, err, author.ErrNotFound)
}

func TestFindByForeignID_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	db.Close()

	_, err := svc.FindByForeignID(ctx, "some-fid")
	require.Error(t, err)
	require.NotErrorIs(t, err, author.ErrNotFound)
}

func TestFindByName_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	db.Close()

	_, err := svc.FindByName(ctx, "Some Author")
	require.Error(t, err)
	require.NotErrorIs(t, err, author.ErrNotFound)
}

func TestList_CountQueryError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	db.Close()

	_, _, err := svc.List(ctx, author.ListAuthorsFilter{})
	require.Error(t, err)
}

func TestCreate_ForeignIDCheckError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	db.Close()

	// Non-empty ForeignID causes FindByForeignID to be called; with DB closed it
	// returns a non-ErrNotFound error triggering "check existing author by foreign_id".
	_, err := svc.Create(ctx, author.CreateAuthorInput{
		Name:      "Test Author",
		ForeignID: "some-fid",
	})
	require.Error(t, err)
}

func TestCreate_NameCheckError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	db.Close()

	// Empty ForeignID skips the ForeignID block entirely; the name check then
	// calls FindByName which fails with a non-ErrNotFound error, triggering the
	// "check existing author by name" error branch.
	_, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Test Author",
	})
	require.Error(t, err)
}

func TestGetStats_DBError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	db.Close()

	_, err := svc.GetStats(ctx, 1)
	require.Error(t, err)
}

func TestList_MonitoredFalseFilter(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Active Author", Monitored: true, Path: t.TempDir()})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Unmonitored Author", Monitored: false, Path: t.TempDir()})

	monitored := false
	authors, total, err := svc.List(ctx, author.ListAuthorsFilter{Monitored: &monitored})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, authors, 1)
	assert.Equal(t, "Unmonitored Author", authors[0].Name)
}

func TestList_SortBySortName(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Zzz Last", SortName: "Last, Zzz", Path: t.TempDir()})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Aaa First", SortName: "First, Aaa", Path: t.TempDir()})

	authors, _, err := svc.List(ctx, author.ListAuthorsFilter{SortBy: "sortName", SortDir: "asc"})
	require.NoError(t, err)
	require.Len(t, authors, 2)
	assert.Equal(t, "Aaa First", authors[0].Name)
}

func TestList_SortByAddedAt(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Author One", Path: t.TempDir()})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Author Two", Path: t.TempDir()})

	// Exercise the "addedAt" sort branch; ordering result is not important here.
	authors, _, err := svc.List(ctx, author.ListAuthorsFilter{SortBy: "addedAt", SortDir: "desc"})
	require.NoError(t, err)
	assert.Len(t, authors, 2)
}

func TestDelete_WithFiles_FolderAbsent(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	dir := t.TempDir()
	testutil.SeedRootFolder(t, db, dir, "test")
	authorID := testutil.SeedAuthor(t, db, "Author Without Folder")

	// Author folder does not exist on disk — deleteAuthorFiles handles it gracefully.
	err := svc.Delete(ctx, authorID, true)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, authorID)
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestDelete_WithFiles_FolderExists(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := author.New(db)
	ctx := context.Background()

	dir := t.TempDir()
	testutil.SeedRootFolder(t, db, dir, "test")
	authorID := testutil.SeedAuthor(t, db, "Author With Books")

	// "Author With Books" contains no special chars so sanitizeFolderName is a no-op.
	authorDir := filepath.Join(dir, "Author With Books")
	require.NoError(t, os.MkdirAll(authorDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(authorDir, "book.epub"), []byte("data"), 0644))

	err := svc.Delete(ctx, authorID, true)
	require.NoError(t, err)

	_, statErr := os.Stat(authorDir)
	assert.True(t, os.IsNotExist(statErr), "author folder should have been removed")
}
