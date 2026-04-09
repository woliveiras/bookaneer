package wanted_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

func TestGetHistory_Empty(t *testing.T) {
	svc, ctx := newTestService(t)

	items, err := svc.GetHistory(ctx, 50, "")
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestGetHistory_WithEvents(t *testing.T) {
	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Author Name")
	bookID := testutil.SeedBook(t, db, authorID, "Test Book")

	_, err := db.Exec(`
		INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
		VALUES (?, ?, 'grabbed', 'Release Title', 'epub', '{"provider":"annas"}')
	`, bookID, authorID)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
		VALUES (?, ?, 'bookImported', 'Test Book', 'epub', '{}')
	`, bookID, authorID)
	require.NoError(t, err)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	items, err := svc.GetHistory(ctx, 50, "")
	require.NoError(t, err)
	assert.Len(t, items, 2)

	assert.Equal(t, "bookImported", items[0].EventType)
	assert.Equal(t, "grabbed", items[1].EventType)
	assert.Equal(t, "Test Book", items[0].BookTitle)
	assert.Equal(t, "Author Name", items[0].AuthorName)
}

func TestGetHistory_FilterByEventType(t *testing.T) {
	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	_, _ = db.Exec(`INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
		VALUES (?, ?, 'grabbed', 'R1', 'epub', '{}')`, bookID, authorID)
	_, _ = db.Exec(`INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
		VALUES (?, ?, 'bookImported', 'R2', 'epub', '{}')`, bookID, authorID)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	items, err := svc.GetHistory(ctx, 50, "grabbed")
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "grabbed", items[0].EventType)
}

func TestGetHistory_Limit(t *testing.T) {
	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	for i := 0; i < 5; i++ {
		_, _ = db.Exec(`INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
			VALUES (?, ?, 'grabbed', 'Release', 'epub', '{}')`, bookID, authorID)
	}

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	items, err := svc.GetHistory(ctx, 3, "")
	require.NoError(t, err)
	assert.Len(t, items, 3)
}

func TestGetBlocklist_Empty(t *testing.T) {
	svc, ctx := newTestService(t)

	items, err := svc.GetBlocklist(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestAddAndRemoveBlocklist(t *testing.T) {
	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	err := svc.AddToBlocklist(ctx, bookID, "Bad Release", "epub", "low quality")
	require.NoError(t, err)

	items, err := svc.GetBlocklist(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Bad Release", items[0].SourceTitle)
	assert.Equal(t, "low quality", items[0].Reason)
	assert.Equal(t, "Book", items[0].BookTitle)
	assert.Equal(t, "Author", items[0].AuthorName)

	err = svc.RemoveFromBlocklist(ctx, items[0].ID)
	require.NoError(t, err)

	items, err = svc.GetBlocklist(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestGetHistory_NullBookAndAuthor(t *testing.T) {
	db := testutil.OpenTestDB(t)

	_, err := db.Exec(`
		INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
		VALUES (NULL, NULL, 'downloadFailed', 'Some Release', 'unknown', '{}')
	`)
	require.NoError(t, err)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	items, err := svc.GetHistory(ctx, 50, "")
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Nil(t, items[0].BookID)
	assert.Nil(t, items[0].AuthorID)
	assert.Equal(t, "", items[0].BookTitle)
}
