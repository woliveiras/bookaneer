package series

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func createTestSeries(t *testing.T, s *Service, title string) *Series {
	t.Helper()
	ctx := context.Background()
	ser, err := s.Create(ctx, CreateSeriesInput{
		ForeignID:   "test-" + title,
		Title:       title,
		Description: "Test description",
		Monitored:   true,
	})
	require.NoError(t, err)
	return ser
}

func TestCreate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	ser, err := s.Create(ctx, CreateSeriesInput{
		ForeignID:   "OL123",
		Title:       "The Lord of the Rings",
		Description: "Epic fantasy",
		Monitored:   true,
	})
	require.NoError(t, err)
	assert.Greater(t, ser.ID, int64(0))
	assert.Equal(t, "The Lord of the Rings", ser.Title)
	assert.True(t, ser.Monitored)
}

func TestCreate_EmptyTitle(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.Create(ctx, CreateSeriesInput{Title: ""})
	assert.ErrorIs(t, err, ErrInvalidInput)
}

func TestCreate_DuplicateForeignID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.Create(ctx, CreateSeriesInput{ForeignID: "dup", Title: "A"})
	require.NoError(t, err)

	_, err = s.Create(ctx, CreateSeriesInput{ForeignID: "dup", Title: "B"})
	assert.ErrorIs(t, err, ErrDuplicate)
}

func TestFindByID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)

	created := createTestSeries(t, s, "Harry Potter")

	got, err := s.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Harry Potter", got.Title)
}

func TestFindByID_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)

	_, err := s.FindByID(context.Background(), 999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestList_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)

	items, total, err := s.List(context.Background(), ListSeriesFilter{})
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
}

func TestList_WithFilter(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	createTestSeries(t, s, "Fantasy Series")
	createTestSeries(t, s, "Sci-Fi Series")

	items, total, err := s.List(ctx, ListSeriesFilter{Search: "Fantasy"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, items, 1)
	assert.Equal(t, "Fantasy Series", items[0].Title)
}

func TestUpdate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	created := createTestSeries(t, s, "Old Title")

	newTitle := "New Title"
	updated, err := s.Update(ctx, created.ID, UpdateSeriesInput{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
}

func TestUpdate_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)

	title := "Ghost"
	_, err := s.Update(context.Background(), 999, UpdateSeriesInput{Title: &title})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	created := createTestSeries(t, s, "ToDelete")

	err := s.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = s.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)

	err := s.Delete(context.Background(), 999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestAddBook_SeriesNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)

	err := s.AddBook(context.Background(), 999, AddBookInput{BookID: 1, Position: "1"})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestRemoveBook_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	created := createTestSeries(t, s, "Empty Series")

	err := s.RemoveBook(ctx, created.ID, 999)
	assert.ErrorIs(t, err, ErrBookNotFound)
}

func TestList_MonitoredFalse(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.Create(ctx, CreateSeriesInput{Title: "Unmonitored", ForeignID: "mon-false-1", Monitored: false})
	require.NoError(t, err)
	_, err = s.Create(ctx, CreateSeriesInput{Title: "Monitored", ForeignID: "mon-false-2", Monitored: true})
	require.NoError(t, err)

	m := false
	items, total, err := s.List(ctx, ListSeriesFilter{Monitored: &m})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.False(t, items[0].Monitored)
}

func TestList_SortByBookCount(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	ctx := context.Background()

	_, err := s.Create(ctx, CreateSeriesInput{Title: "Series BC", ForeignID: "sbc-1"})
	require.NoError(t, err)

	items, _, err := s.List(ctx, ListSeriesFilter{SortBy: "bookCount"})
	require.NoError(t, err)
	assert.NotEmpty(t, items)
}

func TestFindByID_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	db.Close()

	_, err := s.FindByID(context.Background(), 1)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrNotFound)
}

func TestList_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	db.Close()

	_, _, err := s.List(context.Background(), ListSeriesFilter{})
	require.Error(t, err)
}

func TestCreate_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	db.Close()

	_, err := s.Create(context.Background(), CreateSeriesInput{Title: "Test"})
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrDuplicate)
}

func TestDelete_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db)
	db.Close()

	err := s.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrNotFound)
}
