package author_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woliveiras/bookaneer/internal/core/author"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestFindByForeignID(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Tolkien", ForeignID: "OL123A", Monitored: true, Path: t.TempDir(),
	})
	require.NoError(t, err)

	found, err := svc.FindByForeignID(ctx, "OL123A")
	require.NoError(t, err)
	assert.Equal(t, "Tolkien", found.Name)
}

func TestFindByForeignID_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	_, err := svc.FindByForeignID(context.Background(), "nonexistent")
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestFindByName_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	_, err := svc.FindByName(context.Background(), "ghost")
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestCreate_WithSortName(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "J.R.R. Tolkien", SortName: "Tolkien, J.R.R.", Monitored: true, Path: t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, "Tolkien, J.R.R.", created.SortName)
}

func TestUpdate_MultipleFields(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Old", Monitored: true, Path: t.TempDir(),
	})
	require.NoError(t, err)

	newName := "New"
	monitored := false
	overview := "A great author"
	updated, err := svc.Update(ctx, created.ID, author.UpdateAuthorInput{
		Name: &newName, Monitored: &monitored, Overview: &overview,
	})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.False(t, updated.Monitored)
	assert.Equal(t, "A great author", updated.Overview)
}

func TestDelete_WithBooks(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Author", Monitored: true, Path: t.TempDir(),
	})
	require.NoError(t, err)
	testutil.SeedBook(t, db.DB, created.ID, "Book 1")

	err = svc.Delete(ctx, created.ID, false)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, created.ID)
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestList_MonitoredFilter(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Mon1", Monitored: true, Path: t.TempDir()})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Unmon", Monitored: false, Path: t.TempDir()})

	monitored := true
	items, total, err := svc.List(ctx, author.ListAuthorsFilter{Monitored: &monitored})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, items, 1)
	assert.Equal(t, "Mon1", items[0].Name)
}

func TestList_Pagination(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	for i := range 5 {
		_, _ = svc.Create(ctx, author.CreateAuthorInput{
			Name: "Author " + string(rune('A'+i)), Monitored: true, Path: t.TempDir(),
		})
	}

	items, total, err := svc.List(ctx, author.ListAuthorsFilter{Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, items, 2)
}

func TestList_SortBy(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Zebra", Monitored: true, Path: t.TempDir()})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Alpha", Monitored: true, Path: t.TempDir()})

	items, _, err := svc.List(ctx, author.ListAuthorsFilter{SortBy: "name", SortDir: "asc"})
	require.NoError(t, err)
	assert.Equal(t, "Alpha", items[0].Name)
}

func TestCreate_ExistingForeignID_Remonitors(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Author", ForeignID: "OL1A", Monitored: false, Path: t.TempDir(),
	})
	require.NoError(t, err)
	assert.False(t, created.Monitored)

	remon, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Author 2", ForeignID: "OL1A", Monitored: true, Path: t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, remon.ID)
	assert.True(t, remon.Monitored)
}

func TestCreate_ExistingName_Remonitors(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Tolkien", Monitored: false, Path: t.TempDir(),
	})
	require.NoError(t, err)

	remon, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Tolkien", Monitored: true, Path: t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, created.ID, remon.ID)
	assert.True(t, remon.Monitored)
}

func TestCreate_DefaultStatus(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	created, err := svc.Create(context.Background(), author.CreateAuthorInput{
		Name: "Author", Monitored: true, Path: t.TempDir(),
	})
	require.NoError(t, err)
	assert.Equal(t, "active", created.Status)
}

func TestUpdate_AllFields(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{Name: "Old", Path: t.TempDir()})
	require.NoError(t, err)

	name := "New"
	sortName := "New, The"
	foreignID := "OL99A"
	overview := "Bio"
	imageURL := "http://img.jpg"
	status := "ended"
	monitored := true
	path := t.TempDir()

	updated, err := svc.Update(ctx, created.ID, author.UpdateAuthorInput{
		Name: &name, SortName: &sortName, ForeignID: &foreignID,
		Overview: &overview, ImageURL: &imageURL, Status: &status,
		Monitored: &monitored, Path: &path,
	})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, "New, The", updated.SortName)
	assert.Equal(t, "OL99A", updated.ForeignID)
	assert.Equal(t, "Bio", updated.Overview)
	assert.Equal(t, "http://img.jpg", updated.ImageURL)
	assert.Equal(t, "ended", updated.Status)
	assert.True(t, updated.Monitored)
}

func TestUpdate_EmptyInput(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{Name: "Author", Path: t.TempDir()})
	require.NoError(t, err)

	result, err := svc.Update(ctx, created.ID, author.UpdateAuthorInput{})
	require.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
}

func TestUpdate_MissingAuthor(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	name := "X"
	_, err := svc.Update(context.Background(), 9999, author.UpdateAuthorInput{Name: &name})
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestDelete_MissingAuthor(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	err := svc.Delete(context.Background(), 9999, false)
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestList_StatusFilter(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Active", Status: "active", Path: t.TempDir()})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Ended", Status: "ended", Path: t.TempDir()})

	items, total, err := svc.List(ctx, author.ListAuthorsFilter{Status: "ended"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "Ended", items[0].Name)
}

func TestList_SortDesc(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Alpha", Path: t.TempDir()})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Zebra", Path: t.TempDir()})

	items, _, err := svc.List(ctx, author.ListAuthorsFilter{SortBy: "name", SortDir: "desc"})
	require.NoError(t, err)
	assert.Equal(t, "Zebra", items[0].Name)
}

func TestList_Offset(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	for i := range 3 {
		_, _ = svc.Create(ctx, author.CreateAuthorInput{
			Name: "Author " + string(rune('A'+i)), Path: t.TempDir(),
		})
	}

	items, total, err := svc.List(ctx, author.ListAuthorsFilter{Limit: 1, Offset: 1})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, items, 1)
}

func TestGetStats_WithFilesAndMissing(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{Name: "Author", Path: t.TempDir()})
	require.NoError(t, err)

	bookID := testutil.SeedBook(t, db.DB, created.ID, "Book 1")
	testutil.SeedBook(t, db.DB, created.ID, "Book 2")

	_, err = db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/a.epub', 'a.epub', 5000, 'epub', 'epub')`, bookID)
	require.NoError(t, err)

	stats, err := svc.GetStats(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.BookCount)
	assert.Equal(t, 1, stats.BookFileCount)
	assert.Equal(t, 1, stats.MissingBooks)
	assert.Equal(t, 5000, stats.TotalSizeBytes)
}

func TestDelete_WithFiles(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	dir := t.TempDir()
	created, err := svc.Create(ctx, author.CreateAuthorInput{Name: "Author ToDelete Files", Path: dir})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID, true)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, created.ID)
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestUpdate_DuplicateForeignID(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, author.CreateAuthorInput{Name: "A1", ForeignID: "dup-fid-author"})
	require.NoError(t, err)
	a2, err := svc.Create(ctx, author.CreateAuthorInput{Name: "A2", ForeignID: "other-author-fid"})
	require.NoError(t, err)

	fid := "dup-fid-author"
	_, err = svc.Update(ctx, a2.ID, author.UpdateAuthorInput{ForeignID: &fid})
	require.ErrorIs(t, err, author.ErrDuplicate)
}

func TestCreate_WithAllFields(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name:      "Full Author",
		SortName:  "Author, Full",
		ForeignID: "full-author-fid",
		Overview:  "An overview",
		ImageURL:  "http://img.com/photo.jpg",
		Status:    "ended",
		Monitored: true,
		Path:      "/books/full-author",
	})
	require.NoError(t, err)
	assert.Equal(t, "Author, Full", created.SortName)
	assert.Equal(t, "ended", created.Status)
	assert.True(t, created.Monitored)
	assert.Equal(t, "/books/full-author", created.Path)
}

func TestList_SearchByName(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Stephen King", ForeignID: "sbn-sk"})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Jane Austen", ForeignID: "sbn-ja"})

	items, total, err := svc.List(ctx, author.ListAuthorsFilter{Search: "stephen"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, items, 1)
	assert.Equal(t, "Stephen King", items[0].Name)
}
