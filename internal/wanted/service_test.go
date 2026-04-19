package wanted_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/bypass"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

func TestGetBookInfo(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db.DB, bypass.Noop{})
	svc := wanted.New(db.DB, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db.DB, "Tolkien")
	bookID := testutil.SeedBook(t, db.DB, authorID, "The Hobbit")

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
