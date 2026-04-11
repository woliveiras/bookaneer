package opds

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func setupOPDS(t *testing.T) (*echo.Echo, *Server) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	s := New(db)
	e := echo.New()
	s.Register(e)
	return e, s
}

func TestRoot(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "xml")

	var feed Feed
	err := xml.Unmarshal(rec.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Equal(t, "Bookaneer Library", feed.Title)
	assert.Equal(t, "bookaneer:catalog", feed.ID)
	assert.Len(t, feed.Entries, 2) // By Author + Recently Added
	assert.Equal(t, "By Author", feed.Entries[0].Title)
	assert.Equal(t, "Recently Added", feed.Entries[1].Title)
}

func TestRoot_Slash(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthors_EmptyDB(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds/authors", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var feed Feed
	err := xml.Unmarshal(rec.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Equal(t, "Authors", feed.Title)
	assert.Empty(t, feed.Entries)
}

func TestAuthors_WithData(t *testing.T) {
	db := testutil.OpenTestDB(t)
	e := echo.New()
	s := New(db)
	s.Register(e)

	// Seed data
	authorID := testutil.SeedAuthor(t, db, "Machado de Assis")
	bookID := testutil.SeedBook(t, db, authorID, "Dom Casmurro")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality, hash) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		bookID, "/library/test.epub", "Machado de Assis/test.epub", 1024, "epub", "epub", "abc123")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/opds/authors", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var feed Feed
	err = xml.Unmarshal(rec.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Len(t, feed.Entries, 1)
	assert.Equal(t, "Machado de Assis", feed.Entries[0].Title)
}

func TestAuthorBooks(t *testing.T) {
	db := testutil.OpenTestDB(t)
	e := echo.New()
	s := New(db)
	s.Register(e)

	authorID := testutil.SeedAuthor(t, db, "Brandon Sanderson")
	bookID := testutil.SeedBook(t, db, authorID, "The Way of Kings")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality, hash) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		bookID, "/library/book.epub", "Brandon Sanderson/book.epub", 2048, "epub", "epub", "def456")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/opds/authors/1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var feed Feed
	err = xml.Unmarshal(rec.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Equal(t, "Brandon Sanderson", feed.Title)
	assert.Len(t, feed.Entries, 1)
	assert.Equal(t, "The Way of Kings", feed.Entries[0].Title)
	assert.Equal(t, "Brandon Sanderson", feed.Entries[0].Author.Name)

	// Check acquisition link
	assert.NotEmpty(t, feed.Entries[0].Links)
	assert.Equal(t, "http://opds-spec.org/acquisition", feed.Entries[0].Links[0].Rel)
	assert.Equal(t, "application/epub+zip", feed.Entries[0].Links[0].Type)
}

func TestAuthorBooks_NotFound(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds/authors/99999", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAuthorBooks_InvalidID(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds/authors/abc", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRecent_Empty(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds/recent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var feed Feed
	err := xml.Unmarshal(rec.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Equal(t, "Recently Added", feed.Title)
	assert.Empty(t, feed.Entries)
}

func TestSearch_MissingQuery(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds/search", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSearch_WithQuery(t *testing.T) {
	db := testutil.OpenTestDB(t)
	e := echo.New()
	s := New(db)
	s.Register(e)

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")
	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality, hash) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		bookID, "/library/hobbit.epub", "Tolkien/hobbit.epub", 1024, "epub", "epub", "ghi789")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/opds/search?q=Hobbit", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var feed Feed
	err = xml.Unmarshal(rec.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Contains(t, feed.Title, "Hobbit")
	assert.Len(t, feed.Entries, 1)
}

func TestSearch_NoResults(t *testing.T) {
	e, _ := setupOPDS(t)

	req := httptest.NewRequest(http.MethodGet, "/opds/search?q=nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var feed Feed
	err := xml.Unmarshal(rec.Body.Bytes(), &feed)
	require.NoError(t, err)
	assert.Empty(t, feed.Entries)
}

func TestFormatMimeType(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"epub", "application/epub+zip"},
		{"pdf", "application/pdf"},
		{"mobi", "application/x-mobipocket-ebook"},
		{"azw3", "application/vnd.amazon.ebook"},
		{"cbz", "application/x-cbz"},
		{"unknown", "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			assert.Equal(t, tt.want, formatMimeType(tt.format))
		})
	}
}
