// Package opds implements an OPDS 1.2 catalog server.
package opds

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
)

const (
	atomNS         = "http://www.w3.org/2005/Atom"
	opdsNS         = "http://opds-spec.org/2010/catalog"
	searchNS       = "http://a9.com/-/spec/opensearch/1.1/"
	entriesPerPage = 50
)

// Feed is an OPDS Atom feed.
type Feed struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Author  *FeedAuthor `xml:"author,omitempty"`
	Links   []Link      `xml:"link"`
	Entries []Entry     `xml:"entry"`
}

// FeedAuthor is the author of the feed.
type FeedAuthor struct {
	Name string `xml:"name"`
	URI  string `xml:"uri,omitempty"`
}

// Entry is an OPDS Atom entry.
type Entry struct {
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Content string      `xml:"content,omitempty"`
	Author  *FeedAuthor `xml:"author,omitempty"`
	Links   []Link      `xml:"link"`
}

// Link is an Atom link element.
type Link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
}

// Server serves OPDS feeds.
type Server struct {
	db *sql.DB
}

// New creates a new OPDS Server.
func New(db *sql.DB) *Server {
	return &Server{db: db}
}

// Register registers OPDS routes on the given Echo instance.
func (s *Server) Register(e *echo.Echo) {
	opds := e.Group("/opds")
	opds.GET("", s.Root)
	opds.GET("/", s.Root)
	opds.GET("/authors", s.Authors)
	opds.GET("/authors/:id", s.AuthorBooks)
	opds.GET("/recent", s.Recent)
	opds.GET("/search", s.Search)
}

func (s *Server) Root(c *echo.Context) error {
	now := time.Now().UTC().Format(time.RFC3339)
	feed := Feed{
		XMLNS:   atomNS,
		Title:   "Bookaneer Library",
		ID:      "bookaneer:catalog",
		Updated: now,
		Author:  &FeedAuthor{Name: "Bookaneer"},
		Links: []Link{
			{Rel: "self", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "search", Href: "/opds/search?q={searchTerms}", Type: "application/atom+xml"},
		},
		Entries: []Entry{
			{
				Title:   "By Author",
				ID:      "bookaneer:authors",
				Updated: now,
				Content: "Browse books by author",
				Links:   []Link{{Href: "/opds/authors", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"}},
			},
			{
				Title:   "Recently Added",
				ID:      "bookaneer:recent",
				Updated: now,
				Content: "Recently added books",
				Links:   []Link{{Href: "/opds/recent", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
			},
		},
	}
	return c.XML(http.StatusOK, feed)
}

func (s *Server) Authors(c *echo.Context) error {
	ctx := c.Request().Context()
	now := time.Now().UTC().Format(time.RFC3339)

	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, a.name, COUNT(bf.id) as file_count
		FROM authors a
		JOIN books b ON b.author_id = a.id
		JOIN book_files bf ON bf.book_id = b.id
		GROUP BY a.id
		ORDER BY a.name
	`)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer func() { _ = rows.Close() }()

	entries := make([]Entry, 0)
	for rows.Next() {
		var id int64
		var name string
		var fileCount int
		if err := rows.Scan(&id, &name, &fileCount); err != nil {
			continue
		}
		entries = append(entries, Entry{
			Title:   name,
			ID:      fmt.Sprintf("bookaneer:author:%d", id),
			Updated: now,
			Content: fmt.Sprintf("%d book(s)", fileCount),
			Links:   []Link{{Href: fmt.Sprintf("/opds/authors/%d", id), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		})
	}

	feed := Feed{
		XMLNS:   atomNS,
		Title:   "Authors",
		ID:      "bookaneer:authors",
		Updated: now,
		Links:   []Link{{Rel: "self", Href: "/opds/authors", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"}},
		Entries: entries,
	}
	return c.XML(http.StatusOK, feed)
}

func (s *Server) AuthorBooks(c *echo.Context) error {
	ctx := c.Request().Context()
	now := time.Now().UTC().Format(time.RFC3339)

	authorID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid author id")
	}

	var authorName string
	if err := s.db.QueryRowContext(ctx, "SELECT name FROM authors WHERE id = ?", authorID).Scan(&authorName); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "author not found")
	}

	entries, err := s.bookEntriesForQuery(ctx, `
		SELECT b.id, b.title, a.name, bf.id, bf.path, bf.format, bf.size
		FROM books b
		JOIN authors a ON b.author_id = a.id
		JOIN book_files bf ON bf.book_id = b.id
		WHERE b.author_id = ?
		ORDER BY b.title
	`, authorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	feed := Feed{
		XMLNS:   atomNS,
		Title:   authorName,
		ID:      fmt.Sprintf("bookaneer:author:%d", authorID),
		Updated: now,
		Links:   []Link{{Rel: "self", Href: fmt.Sprintf("/opds/authors/%d", authorID), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		Entries: entries,
	}
	return c.XML(http.StatusOK, feed)
}

func (s *Server) Recent(c *echo.Context) error {
	ctx := c.Request().Context()
	now := time.Now().UTC().Format(time.RFC3339)

	entries, err := s.bookEntriesForQuery(ctx, `
		SELECT b.id, b.title, a.name, bf.id, bf.path, bf.format, bf.size
		FROM books b
		JOIN authors a ON b.author_id = a.id
		JOIN book_files bf ON bf.book_id = b.id
		ORDER BY bf.added_at DESC
		LIMIT ?
	`, entriesPerPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	feed := Feed{
		XMLNS:   atomNS,
		Title:   "Recently Added",
		ID:      "bookaneer:recent",
		Updated: now,
		Links:   []Link{{Rel: "self", Href: "/opds/recent", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		Entries: entries,
	}
	return c.XML(http.StatusOK, feed)
}

func (s *Server) Search(c *echo.Context) error {
	ctx := c.Request().Context()
	now := time.Now().UTC().Format(time.RFC3339)
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "q parameter is required")
	}

	searchPattern := "%" + query + "%"
	entries, err := s.bookEntriesForQuery(ctx, `
		SELECT b.id, b.title, a.name, bf.id, bf.path, bf.format, bf.size
		FROM books b
		JOIN authors a ON b.author_id = a.id
		JOIN book_files bf ON bf.book_id = b.id
		WHERE b.title LIKE ? OR a.name LIKE ?
		ORDER BY b.title
		LIMIT ?
	`, searchPattern, searchPattern, entriesPerPage)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	feed := Feed{
		XMLNS:   atomNS,
		Title:   fmt.Sprintf("Search: %s", query),
		ID:      "bookaneer:search",
		Updated: now,
		Links:   []Link{{Rel: "self", Href: fmt.Sprintf("/opds/search?q=%s", query), Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"}},
		Entries: entries,
	}
	return c.XML(http.StatusOK, feed)
}

func formatMimeType(format string) string {
	switch format {
	case "epub":
		return "application/epub+zip"
	case "pdf":
		return "application/pdf"
	case "mobi":
		return "application/x-mobipocket-ebook"
	case "azw3":
		return "application/vnd.amazon.ebook"
	case "cbz":
		return "application/x-cbz"
	default:
		return "application/octet-stream"
	}
}

func (s *Server) bookEntriesForQuery(ctx context.Context, query string, args ...interface{}) ([]Entry, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	entries := make([]Entry, 0)
	for rows.Next() {
		var bookID, fileID int64
		var title, authorName, path, format string
		var size int64
		if err := rows.Scan(&bookID, &title, &authorName, &fileID, &path, &format, &size); err != nil {
			continue
		}
		entries = append(entries, Entry{
			Title:   title,
			ID:      fmt.Sprintf("bookaneer:book:%d", bookID),
			Updated: time.Now().UTC().Format(time.RFC3339),
			Author:  &FeedAuthor{Name: authorName},
			Links: []Link{
				{Rel: "http://opds-spec.org/acquisition", Href: fmt.Sprintf("/api/v1/book/%d/file/%d/download", bookID, fileID), Type: formatMimeType(format)},
			},
		})
	}
	return entries, rows.Err()
}
