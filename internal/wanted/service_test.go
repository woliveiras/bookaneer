package wanted_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

func TestGetWantedBooks_Empty(t *testing.T) {
	svc, ctx := newTestService(t)

	books, err := svc.GetWantedBooks(ctx)
	require.NoError(t, err)
	assert.Empty(t, books)
}

func TestGetWantedBooks_WithMonitoredBooks(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	testutil.SeedBook(t, db, authorID, "The Hobbit")
	testutil.SeedBook(t, db, authorID, "The Lord of the Rings")

	books, err := svc.GetWantedBooks(ctx)
	require.NoError(t, err)
	assert.Len(t, books, 2)
}

func TestGetWantedBooks_ExcludesBooksWithFiles(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	book1 := testutil.SeedBook(t, db, authorID, "The Hobbit")
	testutil.SeedBook(t, db, authorID, "Silmarillion")

	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/lib/hobbit.epub', 'hobbit.epub', 1024, 'epub', 'epub')`, book1)
	require.NoError(t, err)

	books, err := svc.GetWantedBooks(ctx)
	require.NoError(t, err)
	assert.Len(t, books, 1)
	assert.Equal(t, "Silmarillion", books[0].Title)
}

func TestGetBookInfo(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	title, authorName, err := svc.GetBookInfo(ctx, bookID)
	require.NoError(t, err)
	assert.Equal(t, "The Hobbit", title)
	assert.Equal(t, "Tolkien", authorName)
}

func TestGetBookInfo_NotFound(t *testing.T) {
	svc, ctx := newTestService(t)

	_, _, err := svc.GetBookInfo(ctx, 9999)
	require.Error(t, err)
}

func TestProcessDownloads_EmptyQueue(t *testing.T) {
	svc, ctx := newTestService(t)

	result, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Checked)
	assert.Equal(t, 0, result.Completed)
	assert.Equal(t, 0, result.Failed)
}
