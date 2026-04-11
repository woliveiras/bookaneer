package book_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestGetWithEditions_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	b, err := svc.Create(ctx, book.CreateBookInput{AuthorID: authorID, Title: "T", ForeignID: "err-gwe"})
	require.NoError(t, err)

	_ = db.Close()

	_, err = svc.GetWithEditions(ctx, b.ID)
	require.Error(t, err)
}

func TestCreateEdition_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_ = db.Close()

	_, err := svc.CreateEdition(ctx, book.CreateEditionInput{BookID: 1, Title: "Edition"})
	require.Error(t, err)
}

func TestDeleteEdition_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_ = db.Close()

	err := svc.DeleteEdition(ctx, 1)
	require.Error(t, err)
}

func TestFindByForeignID_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_ = db.Close()

	_, err := svc.FindByForeignID(ctx, "some-foreign-id")
	require.Error(t, err)
	require.NotErrorIs(t, err, book.ErrNotFound)
}

func TestList_CountQueryError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_ = db.Close()

	_, _, err := svc.List(ctx, book.ListBooksFilter{})
	require.Error(t, err)
}

func TestCreate_ForeignIDCheckError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_ = db.Close()

	// Non-empty ForeignID causes FindByForeignID to be called; with DB closed it
	// returns a non-ErrNotFound error which triggers the "check existing book" branch.
	_, err := svc.Create(ctx, book.CreateBookInput{
		AuthorID:  1,
		Title:     "T",
		ForeignID: "fid-check-err",
	})
	require.Error(t, err)
}

func TestDelete_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := book.New(db)
	ctx := context.Background()

	_ = db.Close()

	// book.Delete calls ExecContext directly (no FindByID first), so a closed DB
	// immediately surfaces the ExecContext error branch.
	err := svc.Delete(ctx, 1)
	require.Error(t, err)
}
