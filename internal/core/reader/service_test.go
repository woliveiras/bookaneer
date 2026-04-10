package reader_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/reader"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestNew(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	assert.NotNil(t, svc)
}

func seedBookFile(t *testing.T, db interface {
	Exec(query string, args ...any) (interface{ LastInsertId() (int64, error) }, error)
}, bookID int64) int64 {
	t.Helper()
	// Use raw SQL to insert a book_file directly
	return 0
}

func TestGetBookFile(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Test Author")
	bookID := testutil.SeedBook(t, db, authorID, "Test Book")

	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/test.epub', 'test.epub', 1024, 'epub', 'epub')`, bookID)
	require.NoError(t, err)

	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))

	file, err := svc.GetBookFile(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, fileID, file.ID)
	assert.Equal(t, "epub", file.Format)
	assert.Equal(t, int64(1024), file.Size)
}

func TestGetBookFile_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	_, err := svc.GetBookFile(context.Background(), 9999)
	require.ErrorIs(t, err, reader.ErrBookFileNotFound)
}

func TestListBookFiles(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/a.epub', 'a.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/b.mobi', 'b.mobi', 200, 'mobi', 'mobi')`, bookID)
	require.NoError(t, err)

	files, err := svc.ListBookFiles(ctx, bookID)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestListBookFiles_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	files, err := svc.ListBookFiles(context.Background(), 9999)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestSaveProgress(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/r.epub', 'r.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))

	// Insert a user for FK constraint
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	progress, err := svc.SaveProgress(ctx, fileID, 1, "epubcfi(/6/4)", 0.25)
	require.NoError(t, err)
	assert.Equal(t, "epubcfi(/6/4)", progress.Position)
	assert.Equal(t, 0.25, progress.Percentage)
}

func TestSaveProgress_Update(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/u.epub', 'u.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	_, err = svc.SaveProgress(ctx, fileID, 1, "pos1", 0.1)
	require.NoError(t, err)

	// Updating same file/user
	progress, err := svc.SaveProgress(ctx, fileID, 1, "pos2", 0.5)
	require.NoError(t, err)
	assert.Equal(t, "pos2", progress.Position)
	assert.Equal(t, 0.5, progress.Percentage)
}

func TestGetProgress(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/g.epub', 'g.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	_, err = svc.SaveProgress(ctx, fileID, 1, "pos", 0.75)
	require.NoError(t, err)

	progress, err := svc.GetProgress(ctx, fileID, 1)
	require.NoError(t, err)
	assert.Equal(t, "pos", progress.Position)
	assert.Equal(t, 0.75, progress.Percentage)
}

func TestGetProgress_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	_, err := svc.GetProgress(context.Background(), 9999, 1)
	require.ErrorIs(t, err, reader.ErrProgressNotFound)
}

func TestCreateBookmark(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/bm.epub', 'bm.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	bm, err := svc.CreateBookmark(ctx, fileID, 1, "epubcfi(/6/4)", "Chapter 1", "Great chapter!")
	require.NoError(t, err)
	assert.NotZero(t, bm.ID)
	assert.Equal(t, "Chapter 1", bm.Title)
	assert.Equal(t, "Great chapter!", bm.Note)
}

func TestListBookmarks(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/lb.epub', 'lb.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	_, err = svc.CreateBookmark(ctx, fileID, 1, "pos1", "BM1", "")
	require.NoError(t, err)
	_, err = svc.CreateBookmark(ctx, fileID, 1, "pos2", "BM2", "note2")
	require.NoError(t, err)

	bms, err := svc.ListBookmarks(ctx, fileID, 1)
	require.NoError(t, err)
	assert.Len(t, bms, 2)
}

func TestListBookmarks_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	bms, err := svc.ListBookmarks(context.Background(), 9999, 1)
	require.NoError(t, err)
	assert.Empty(t, bms)
}

func TestGetBookmark(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/gb.epub', 'gb.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	created, err := svc.CreateBookmark(ctx, fileID, 1, "pos", "Title", "Note")
	require.NoError(t, err)

	bm, err := svc.GetBookmark(ctx, created.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "Title", bm.Title)
}

func TestGetBookmark_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	_, err := svc.GetBookmark(context.Background(), 9999, 1)
	require.ErrorIs(t, err, reader.ErrBookmarkNotFound)
}

func TestDeleteBookmark(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/db.epub', 'db.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	var fileID int64
	require.NoError(t, db.QueryRow("SELECT id FROM book_files WHERE book_id = ?", bookID).Scan(&fileID))
	_, err = db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	created, err := svc.CreateBookmark(ctx, fileID, 1, "pos", "Title", "")
	require.NoError(t, err)

	err = svc.DeleteBookmark(ctx, created.ID, 1)
	require.NoError(t, err)

	_, err = svc.GetBookmark(ctx, created.ID, 1)
	require.ErrorIs(t, err, reader.ErrBookmarkNotFound)
}

func TestDeleteBookmark_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	err := svc.DeleteBookmark(context.Background(), 9999, 1)
	require.ErrorIs(t, err, reader.ErrBookmarkNotFound)
}

func TestSaveProgress_InvalidBookFile(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	_, err = svc.SaveProgress(ctx, 99999, 1, "epubcfi(/6/2)", 10.0)
	require.Error(t, err)
}

func TestCreateBookmark_InvalidBookFile(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO users (id, username, password_hash, api_key) VALUES (1, 'test', 'hash', 'test-api-key-' || hex(randomblob(8)))`)
	require.NoError(t, err)

	_, err = svc.CreateBookmark(ctx, 99999, 1, "epubcfi(/6/4)", "Test", "Note")
	require.Error(t, err)
}

func TestListBookFiles_WithMultipleFiles(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := reader.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Multi Author")
	bookID := testutil.SeedBook(t, db, authorID, "Multi Book")

	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/a.epub', 'a.epub', 100, 'epub', 'epub')`, bookID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/b.pdf', 'b.pdf', 200, 'pdf', 'pdf')`, bookID)
	require.NoError(t, err)

	files, err := svc.ListBookFiles(ctx, bookID)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}
