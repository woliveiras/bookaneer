package author_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woliveiras/bookaneer/internal/core/author"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func newTestService(t *testing.T) *author.Service {
	t.Helper()
	return author.New(testutil.OpenTestDBX(t))
}

func TestCreate_Success(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	a, err := svc.Create(ctx, author.CreateAuthorInput{
		Name:      "J.R.R. Tolkien",
		SortName:  "Tolkien, J.R.R.",
		ForeignID: "OL26320A",
		Monitored: true,
		Path:      "/library/J.R.R. Tolkien",
	})
	require.NoError(t, err)
	assert.NotZero(t, a.ID)
	assert.Equal(t, "J.R.R. Tolkien", a.Name)
	assert.Equal(t, "Tolkien, J.R.R.", a.SortName)
	assert.True(t, a.Monitored)
}

func TestCreate_EmptyName(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, author.CreateAuthorInput{})
	require.ErrorIs(t, err, author.ErrInvalidInput)
}

func TestCreate_DuplicateForeignID(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, author.CreateAuthorInput{
		Name:      "Tolkien",
		ForeignID: "OL26320A",
		Path:      "/library/Tolkien",
	})
	require.NoError(t, err)

	// Creating again with same foreign ID should return existing (updated)
	a2, err := svc.Create(ctx, author.CreateAuthorInput{
		Name:      "J.R.R. Tolkien",
		ForeignID: "OL26320A",
		Path:      "/library/J.R.R. Tolkien",
	})
	require.NoError(t, err)
	assert.True(t, a2.Monitored) // Gets set to true
}

func TestCreate_DuplicateName(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Tolkien",
		Path: "/library/Tolkien",
	})
	require.NoError(t, err)

	// Same name should return existing
	a2, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Tolkien",
		Path: "/library/Tolkien2",
	})
	require.NoError(t, err)
	assert.True(t, a2.Monitored)
}

func TestFindByID(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name:      "Brandon Sanderson",
		Monitored: true,
		Path:      "/library/Brandon Sanderson",
	})
	require.NoError(t, err)

	found, err := svc.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Brandon Sanderson", found.Name)
}

func TestFindByID_NotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, err := svc.FindByID(ctx, 9999)
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestFindByName(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "Patrick Rothfuss",
		Path: "/library/Patrick Rothfuss",
	})
	require.NoError(t, err)

	// Case-insensitive
	found, err := svc.FindByName(ctx, "patrick rothfuss")
	require.NoError(t, err)
	assert.Equal(t, "Patrick Rothfuss", found.Name)
}

func TestUpdate(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "NK Jemisin",
		Path: "/library/NK Jemisin",
	})
	require.NoError(t, err)

	newName := "N.K. Jemisin"
	updated, err := svc.Update(ctx, created.ID, author.UpdateAuthorInput{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "N.K. Jemisin", updated.Name)
}

func TestUpdate_NotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, err := svc.Update(ctx, 9999, author.UpdateAuthorInput{})
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestDelete(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, author.CreateAuthorInput{
		Name: "To Delete",
		Path: "/library/To Delete",
	})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID, false)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, created.ID)
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	err := svc.Delete(ctx, 9999, false)
	require.ErrorIs(t, err, author.ErrNotFound)
}

func TestList_Empty(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	authors, total, err := svc.List(ctx, author.ListAuthorsFilter{})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, authors)
}

func TestList_WithFilter(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Author A", Monitored: true, Path: "/library/A"})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Author B", Monitored: false, Path: "/library/B"})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Author C", Monitored: true, Path: "/library/C"})

	monitored := true
	authors, total, err := svc.List(ctx, author.ListAuthorsFilter{Monitored: &monitored})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, authors, 2)
}

func TestList_Search(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "J.R.R. Tolkien", Path: "/library/Tolkien"})
	_, _ = svc.Create(ctx, author.CreateAuthorInput{Name: "Brandon Sanderson", Path: "/library/Sanderson"})

	authors, total, err := svc.List(ctx, author.ListAuthorsFilter{Search: "tolkien"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "J.R.R. Tolkien", authors[0].Name)
}

func TestGetStats(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := author.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Tolkien")
	testutil.SeedBook(t, db.DB, authorID, "The Hobbit")
	testutil.SeedBook(t, db.DB, authorID, "LOTR")

	stats, err := svc.GetStats(ctx, authorID)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.BookCount)
	assert.Equal(t, 2, stats.MissingBooks) // No files exist
	assert.Equal(t, 0, stats.BookFileCount)
}
