package series_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/series"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestGetWithBooks_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Empty Series"})
	require.NoError(t, err)

	swb, err := svc.GetWithBooks(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Empty Series", swb.Title)
	assert.Empty(t, swb.Books)
}

func TestGetWithBooks_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	_, err := svc.GetWithBooks(context.Background(), 9999)
	require.ErrorIs(t, err, series.ErrNotFound)
}

func TestAddBook_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Test Series"})
	require.NoError(t, err)

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book 1")

	err = svc.AddBook(ctx, s.ID, series.AddBookInput{BookID: bookID, Position: "1"})
	require.NoError(t, err)

	swb, err := svc.GetWithBooks(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, swb.Books, 1)
	assert.Equal(t, bookID, swb.Books[0].BookID)
	assert.Equal(t, "1", swb.Books[0].Position)
}

func TestAddBook_BookNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Series"})
	require.NoError(t, err)

	err = svc.AddBook(ctx, s.ID, series.AddBookInput{BookID: 9999, Position: "1"})
	require.ErrorIs(t, err, series.ErrBookNotFound)
}

func TestAddBook_AlreadyInSeries(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Series"})
	require.NoError(t, err)

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	err = svc.AddBook(ctx, s.ID, series.AddBookInput{BookID: bookID, Position: "1"})
	require.NoError(t, err)

	err = svc.AddBook(ctx, s.ID, series.AddBookInput{BookID: bookID, Position: "2"})
	require.ErrorIs(t, err, series.ErrBookAlreadyInSeries)
}

func TestRemoveBook_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Series"})
	require.NoError(t, err)

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	err = svc.AddBook(ctx, s.ID, series.AddBookInput{BookID: bookID, Position: "1"})
	require.NoError(t, err)

	err = svc.RemoveBook(ctx, s.ID, bookID)
	require.NoError(t, err)

	swb, err := svc.GetWithBooks(ctx, s.ID)
	require.NoError(t, err)
	assert.Empty(t, swb.Books)
}

func TestUpdateBookPosition_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Series"})
	require.NoError(t, err)

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	err = svc.AddBook(ctx, s.ID, series.AddBookInput{BookID: bookID, Position: "1"})
	require.NoError(t, err)

	err = svc.UpdateBookPosition(ctx, s.ID, bookID, "3")
	require.NoError(t, err)

	swb, err := svc.GetWithBooks(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "3", swb.Books[0].Position)
}

func TestUpdateBookPosition_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Series"})
	require.NoError(t, err)

	err = svc.UpdateBookPosition(ctx, s.ID, 9999, "1")
	require.ErrorIs(t, err, series.ErrBookNotFound)
}

func TestUpdate_Fields(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Old", Description: "Desc", Monitored: true})
	require.NoError(t, err)

	newTitle := "New Title"
	newDesc := "New Desc"
	monitored := false
	updated, err := svc.Update(ctx, s.ID, series.UpdateSeriesInput{Title: &newTitle, Description: &newDesc, Monitored: &monitored})
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
	assert.Equal(t, "New Desc", updated.Description)
	assert.False(t, updated.Monitored)
}

func TestList_Pagination(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	titles := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	for i, title := range titles {
		_, err := svc.Create(ctx, series.CreateSeriesInput{Title: title, ForeignID: "pag-" + string(rune('A'+i))})
		require.NoError(t, err)
	}

	items, total, err := svc.List(ctx, series.ListSeriesFilter{Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, items, 2)
}

func TestList_Search(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, series.CreateSeriesInput{Title: "Lord of the Rings"})
	_, _ = svc.Create(ctx, series.CreateSeriesInput{Title: "Harry Potter"})

	items, total, err := svc.List(ctx, series.ListSeriesFilter{Search: "lord"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, items, 1)
	assert.Equal(t, "Lord of the Rings", items[0].Title)
}

func TestUpdate_DuplicateForeignID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Series A", ForeignID: "dup-fid"})
	require.NoError(t, err)
	s2, err := svc.Create(ctx, series.CreateSeriesInput{Title: "Series B", ForeignID: "dup-fid-b"})
	require.NoError(t, err)

	fid := "dup-fid"
	_, err = svc.Update(ctx, s2.ID, series.UpdateSeriesInput{ForeignID: &fid})
	require.ErrorIs(t, err, series.ErrDuplicate)
}

func TestUpdate_EmptySetsReturnsExisting(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	s, err := svc.Create(ctx, series.CreateSeriesInput{Title: "NoOp"})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, s.ID, series.UpdateSeriesInput{})
	require.NoError(t, err)
	assert.Equal(t, s.Title, updated.Title)
}

func TestList_SortDesc(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, series.CreateSeriesInput{Title: "AAA", ForeignID: "sd-a"})
	_, _ = svc.Create(ctx, series.CreateSeriesInput{Title: "ZZZ", ForeignID: "sd-z"})

	items, _, err := svc.List(ctx, series.ListSeriesFilter{SortBy: "title", SortDir: "desc", Limit: 1})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "ZZZ", items[0].Title)
}

func TestList_Offset(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	for i := range 5 {
		_, _ = svc.Create(ctx, series.CreateSeriesInput{Title: "Off-" + string(rune('A'+i)), ForeignID: "off-" + string(rune('A'+i))})
	}

	items, _, err := svc.List(ctx, series.ListSeriesFilter{Offset: 3, Limit: 10})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(items), 2)
}

func TestList_MonitoredFilter(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := series.New(db)
	ctx := context.Background()

	_, _ = svc.Create(ctx, series.CreateSeriesInput{Title: "Mon", ForeignID: "mon-1", Monitored: true})
	_, _ = svc.Create(ctx, series.CreateSeriesInput{Title: "Unmon", ForeignID: "mon-2", Monitored: false})

	m := true
	items, _, err := svc.List(ctx, series.ListSeriesFilter{Monitored: &m})
	require.NoError(t, err)
	for _, item := range items {
		assert.True(t, item.Monitored)
	}
}
