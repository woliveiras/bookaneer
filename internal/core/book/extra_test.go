package book_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestFindByForeignID_Success(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Foreign Book", ForeignID: "OL123W"})
	require.NoError(t, err)

	found, err := svc.FindByForeignID(ctx, "OL123W")
	require.NoError(t, err)
	assert.Equal(t, "Foreign Book", found.Title)
}

func TestFindByForeignID_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	_, err := svc.FindByForeignID(context.Background(), "nonexistent")
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestList_SortByTitle(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Zebra Book", ForeignID: "sort-z"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Alpha Book", ForeignID: "sort-a"})
	require.NoError(t, err)

	books, total, err := svc.List(ctx, book.ListBooksFilter{})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, books, 2)
	// Default sort is by title ASC; verify both books returned
	titles := []string{books[0].Title, books[1].Title}
	assert.Contains(t, titles, "Zebra Book")
	assert.Contains(t, titles, "Alpha Book")
}

func TestList_SortByReleaseDate(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Old", ReleaseDate: "2000-01-01", ForeignID: "date-old"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "New", ReleaseDate: "2024-01-01", ForeignID: "date-new"})
	require.NoError(t, err)

	books, total, err := svc.List(ctx, book.ListBooksFilter{SortBy: "releaseDate"})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, books, 2)
}

func TestList_SearchFilter(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "The Hobbit", ForeignID: "search-hobbit"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Silmarillion", ForeignID: "search-silm"})
	require.NoError(t, err)

	books, total, err := svc.List(ctx, book.ListBooksFilter{Search: "hobbit"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, books, 1)
	assert.Equal(t, "The Hobbit", books[0].Title)
}

func TestList_MissingFilter(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	b, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Has File", ForeignID: "miss-has"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "No File", ForeignID: "miss-no"})
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/a.epub', 'a.epub', 100, 'epub', 'epub')`, b.ID)
	require.NoError(t, err)

	books, total, err := svc.List(ctx, book.ListBooksFilter{Missing: true})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, books, 1)
	assert.Equal(t, "No File", books[0].Title)
}

func TestUpdate_InWishlist(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Test", InWishlist: true})
	require.NoError(t, err)

	inWishlist := false
	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{InWishlist: &inWishlist})
	require.NoError(t, err)
	assert.False(t, updated.InWishlist)
}

func TestUpdate_Title(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Old"})
	require.NoError(t, err)

	newTitle := "New Title"
	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{Title: &newTitle})
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
}

func TestUpdate_ForeignID(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book"})
	require.NoError(t, err)

	fid := "OL999W"
	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{ForeignID: &fid})
	require.NoError(t, err)
	assert.Equal(t, "OL999W", updated.ForeignID)
}

func TestUpdate_ChangeAuthor(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	author1 := testutil.SeedAuthor(t, db.DB, "Author1")
	author2 := testutil.SeedAuthor(t, db.DB, "Author2")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: author1, Title: "Book"})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{AuthorID: &author2})
	require.NoError(t, err)
	assert.Equal(t, author2, updated.AuthorID)
}

func TestUpdate_AuthorNotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book"})
	require.NoError(t, err)

	badAuthor := int64(9999)
	_, err = svc.Update(ctx, created.ID, book.UpdateBookInput{AuthorID: &badAuthor})
	require.ErrorIs(t, err, book.ErrAuthorNotFound)
}

func TestUpdate_NoChanges(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book"})
	require.NoError(t, err)

	result, err := svc.Update(ctx, created.ID, book.UpdateBookInput{})
	require.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
}

func TestCreateEdition_WithBook(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	b, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book"})
	require.NoError(t, err)

	e, err := svc.CreateEdition(ctx, book.CreateEditionInput{
		BookID: b.ID, Title: "Hardcover", Format: "hardcover", ISBN: "123456",
	})
	require.NoError(t, err)
	assert.Equal(t, "Hardcover", e.Title)
	assert.Equal(t, "hardcover", e.Format)
}

func TestCreateEdition_MissingTitle(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	b, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book"})
	require.NoError(t, err)

	_, err = svc.CreateEdition(ctx, book.CreateEditionInput{BookID: b.ID})
	require.ErrorIs(t, err, book.ErrInvalidInput)
}

func TestGetWithEditions_WithFiles(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	b, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book"})
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/tmp/b.epub', 'b.epub', 1024, 'epub', 'epub')`, b.ID)
	require.NoError(t, err)

	bwe, err := svc.GetWithEditions(ctx, b.ID)
	require.NoError(t, err)
	assert.Len(t, bwe.Files, 1)
	assert.Equal(t, "epub", bwe.Files[0].Format)
}

func TestCreate_ExistingForeignID_UpdatesWishlist(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book", ForeignID: "OL1W", InWishlist: false})
	require.NoError(t, err)
	assert.False(t, created.InWishlist)

	addedToWishlist, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book 2", ForeignID: "OL1W", InWishlist: true})
	require.NoError(t, err)
	assert.Equal(t, created.ID, addedToWishlist.ID)
	assert.True(t, addedToWishlist.InWishlist)
}

func TestList_InWishlistFilter(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Wishlisted", InWishlist: true, ForeignID: "wl-yes"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Not Wishlisted", InWishlist: false, ForeignID: "wl-no"})
	require.NoError(t, err)

	books, total, err := svc.List(ctx, book.ListBooksFilter{InWishlist: true})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, "Wishlisted", books[0].Title)
}

func TestList_Pagination(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	names := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	for i, name := range names {
		_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book " + name, ForeignID: "pag-" + string(rune('A'+i))})
		require.NoError(t, err)
	}

	books, total, err := svc.List(ctx, book.ListBooksFilter{Limit: 2})
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, books, 2)
}

func TestList_Offset(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	for i := range 3 {
		_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book " + string(rune('A'+i)), ForeignID: "off-" + string(rune('A'+i))})
		require.NoError(t, err)
	}

	books, total, err := svc.List(ctx, book.ListBooksFilter{Limit: 1, Offset: 1})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, books, 1)
}

func TestDelete_BookRemoves(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "To Delete", ForeignID: "del-1"})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, created.ID)
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestDelete_BookNotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	err := svc.Delete(context.Background(), 9999)
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestDeleteEdition_Exists(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	b, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book", ForeignID: "ed-del"})
	require.NoError(t, err)

	e, err := svc.CreateEdition(ctx, book.CreateEditionInput{BookID: b.ID, Title: "Edition"})
	require.NoError(t, err)

	err = svc.DeleteEdition(ctx, e.ID)
	require.NoError(t, err)
}

func TestDeleteEdition_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	err := svc.DeleteEdition(context.Background(), 9999)
	require.ErrorIs(t, err, book.ErrEditionNotFound)
}

func TestCreateEdition_BookNotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	_, err := svc.CreateEdition(context.Background(), book.CreateEditionInput{BookID: 9999, Title: "Ed"})
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestCreateEdition_WithAllFields(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	b, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Book", ForeignID: "ed-all"})
	require.NoError(t, err)

	e, err := svc.CreateEdition(ctx, book.CreateEditionInput{
		BookID: b.ID, Title: "Paperback", ForeignID: "OL-PB1", ISBN: "1234567890",
		Format: "paperback", Publisher: "Penguin", ReleaseDate: "2024-01-01",
		PageCount: 300, Language: "en",
	})
	require.NoError(t, err)
	assert.Equal(t, "Paperback", e.Title)
	assert.Equal(t, "OL-PB1", e.ForeignID)
	assert.Equal(t, "paperback", e.Format)
}

func TestUpdate_AllBookFields(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Old", ForeignID: "upd-all"})
	require.NoError(t, err)

	newTitle := "New Title"
	newForeignID := "upd-new"
	inWishlist := true
	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{Title: &newTitle, ForeignID: &newForeignID, InWishlist: &inWishlist})
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
	assert.Equal(t, "upd-new", updated.ForeignID)
	assert.True(t, updated.InWishlist)
}

func TestUpdate_DuplicateForeignID(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "A", ForeignID: "dup-fid-book"})
	require.NoError(t, err)
	b2, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "B", ForeignID: "other-book-fid"})
	require.NoError(t, err)

	fid := "dup-fid-book"
	_, err = svc.Update(ctx, b2.ID, book.UpdateBookInput{ForeignID: &fid})
	require.ErrorIs(t, err, book.ErrDuplicate)
}

func TestUpdate_RemoveFromWishlistCleansQueue(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Mon", ForeignID: "unmon-q", InWishlist: true})
	require.NoError(t, err)

	// Add a queued download
	_, err = db.ExecContext(ctx, `INSERT INTO download_queue (book_id, title, status) VALUES (?, 'test', 'queued')`, created.ID)
	require.NoError(t, err)

	inWishlist := false
	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{InWishlist: &inWishlist})
	require.NoError(t, err)
	assert.False(t, updated.InWishlist)
}

func TestUpdate_MultipleFields(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "Multi", ForeignID: "multi-upd"})
	require.NoError(t, err)

	isbn := "1234567890"
	isbn13 := "1234567890123"
	releaseDate := "2024-06-01"
	overview := "Updated overview"
	imageURL := "http://img.com/cover.jpg"
	pageCount := 300
	sortTitle := "Updated Sort"

	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{
		ISBN:        &isbn,
		ISBN13:      &isbn13,
		ReleaseDate: &releaseDate,
		Overview:    &overview,
		ImageURL:    &imageURL,
		PageCount:   &pageCount,
		SortTitle:   &sortTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, isbn, updated.ISBN)
	assert.Equal(t, isbn13, updated.ISBN13)
	assert.Equal(t, overview, updated.Overview)
	assert.Equal(t, pageCount, updated.PageCount)
}

func TestList_SortDesc(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Author")
	_, _ = svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "AAA Book", ForeignID: "sd-a"})
	_, _ = svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "ZZZ Book", ForeignID: "sd-z"})

	books, _, err := svc.List(ctx, book.ListBooksFilter{SortBy: "title", SortDir: "desc", Limit: 1})
	require.NoError(t, err)
	require.Len(t, books, 1)
	assert.Equal(t, "ZZZ Book", books[0].Title)
}
