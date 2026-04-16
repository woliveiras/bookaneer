package book_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestCreate_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")

	b, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "The Hobbit",
	})
	require.NoError(t, err)
	assert.NotZero(t, b.ID)
	assert.Equal(t, "The Hobbit", b.Title)
	assert.Equal(t, "The Hobbit", b.SortTitle) // Defaults to title
	assert.False(t, b.InWishlist)
	assert.Equal(t, "Tolkien", b.AuthorName)
}

func TestCreate_EmptyTitle(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: 1})
	require.ErrorIs(t, err, book.ErrInvalidInput)
}

func TestCreate_NoAuthorID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, book.CreateBookInput{Title: "Test"})
	require.ErrorIs(t, err, book.ErrInvalidInput)
}

func TestCreate_AuthorNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, book.CreateBookInput{AuthorID: 9999, Title: "Test"})
	require.ErrorIs(t, err, book.ErrAuthorNotFound)
}

func TestCreate_DuplicateForeignID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")

	_, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID:  authorID,
		Title:     "The Hobbit",
		ForeignID: "OL27516W",
	})
	require.NoError(t, err)

	// Same foreign ID returns existing (updated to in_wishlist)
	b2, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID:   authorID,
		Title:      "The Hobbit 2nd",
		ForeignID:  "OL27516W",
		InWishlist: true,
	})
	require.NoError(t, err)
	assert.True(t, b2.InWishlist)
}

func TestFindByID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Sanderson")
	created, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "Mistborn",
		ISBN13:   "9780765311788",
	})
	require.NoError(t, err)

	found, err := svc.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Mistborn", found.Title)
	assert.Equal(t, "9780765311788", found.ISBN13)
	assert.Equal(t, "Sanderson", found.AuthorName)
	assert.False(t, found.HasFile)
}

func TestFindByID_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_, err := svc.FindByID(ctx, 9999)
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestUpdate(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "Original Title",
	})
	require.NoError(t, err)

	newTitle := "Updated Title"
	updated, err := svc.Update(ctx, created.ID, book.UpdateBookInput{
		Title: &newTitle,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
}

func TestUpdate_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_, err := svc.Update(ctx, 9999, book.UpdateBookInput{})
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestDelete(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	created, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "To Delete",
	})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, created.ID)
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	err := svc.Delete(ctx, 9999)
	require.ErrorIs(t, err, book.ErrNotFound)
}

func TestList_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	books, total, err := svc.List(ctx, book.ListBooksFilter{})
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, books)
}

func TestList_ByAuthor(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	a1 := testutil.SeedAuthor(t, db, "Author A")
	a2 := testutil.SeedAuthor(t, db, "Author B")

	_, _ = svc.Create(ctx, book.CreateBookInput{AuthorID: a1, Title: "Book A1", ForeignID: "fid-a1"})
	_, _ = svc.Create(ctx, book.CreateBookInput{AuthorID: a1, Title: "Book A2", ForeignID: "fid-a2"})
	_, _ = svc.Create(ctx, book.CreateBookInput{AuthorID: a2, Title: "Book B1", ForeignID: "fid-b1"})

	books, total, err := svc.List(ctx, book.ListBooksFilter{AuthorID: &a1})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, books, 2)
}

func TestGetWithEditions(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	created, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "The Hobbit",
	})
	require.NoError(t, err)

	bwe, err := svc.GetWithEditions(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "The Hobbit", bwe.Title)
	assert.Empty(t, bwe.Editions)
	assert.Empty(t, bwe.Files)
}

func TestCreateEdition(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	created, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "The Hobbit",
	})
	require.NoError(t, err)

	edition, err := svc.CreateEdition(ctx, book.CreateEditionInput{
		BookID:   created.ID,
		Title:    "The Hobbit - Kindle",
		Format:   "epub",
		Language: "en",
	})
	require.NoError(t, err)
	assert.NotZero(t, edition.ID)
	assert.Equal(t, "The Hobbit - Kindle", edition.Title)
	assert.Equal(t, "epub", edition.Format)

	// Verify via GetWithEditions
	bwe, err := svc.GetWithEditions(ctx, created.ID)
	require.NoError(t, err)
	require.Len(t, bwe.Editions, 1)
	assert.Equal(t, "The Hobbit - Kindle", bwe.Editions[0].Title)
}

func TestDeleteEdition(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	created, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "The Hobbit",
	})
	require.NoError(t, err)

	edition, err := svc.CreateEdition(ctx, book.CreateEditionInput{
		BookID: created.ID,
		Title:  "Edition",
	})
	require.NoError(t, err)

	err = svc.DeleteEdition(ctx, edition.ID)
	require.NoError(t, err)

	err = svc.DeleteEdition(ctx, edition.ID)
	require.ErrorIs(t, err, book.ErrEditionNotFound)
}

func TestDeleteAuthor_CascadesBooks(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	b, err := bookSvc.Create(ctx, book.CreateBookInput{
		AuthorID: authorID,
		Title:    "The Hobbit",
	})
	require.NoError(t, err)

	// Delete author directly via SQL (simulating cascade)
	_, err = db.Exec("DELETE FROM authors WHERE id = ?", authorID)
	require.NoError(t, err)

	// Book should be gone (cascade delete)
	_, err = bookSvc.FindByID(ctx, b.ID)
	require.ErrorIs(t, err, book.ErrNotFound)
}
