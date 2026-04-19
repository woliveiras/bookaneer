package pathmapping_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woliveiras/bookaneer/internal/core/pathmapping"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestNew(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	assert.NotNil(t, svc)
}

func TestCreate_Success(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	m, err := svc.Create(ctx, pathmapping.CreateInput{
		Host:       "sabnzbd",
		RemotePath: "/remote/downloads",
		LocalPath:  "/local/downloads",
	})
	require.NoError(t, err)
	assert.NotZero(t, m.ID)
	assert.Equal(t, "sabnzbd", m.Host)
	assert.Equal(t, "/remote/downloads/", m.RemotePath) // normalized
	assert.Equal(t, "/local/downloads/", m.LocalPath)   // normalized
	assert.NotEmpty(t, m.CreatedAt)
}

func TestCreate_EmptyRemotePath(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)

	_, err := svc.Create(context.Background(), pathmapping.CreateInput{
		Host:      "host",
		LocalPath: "/local/",
	})
	require.ErrorIs(t, err, pathmapping.ErrInvalidInput)
}

func TestCreate_EmptyLocalPath(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)

	_, err := svc.Create(context.Background(), pathmapping.CreateInput{
		Host:       "host",
		RemotePath: "/remote/",
	})
	require.ErrorIs(t, err, pathmapping.ErrInvalidInput)
}

func TestFindByID(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, pathmapping.CreateInput{
		Host: "host", RemotePath: "/remote/", LocalPath: "/local/",
	})
	require.NoError(t, err)

	found, err := svc.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "host", found.Host)
}

func TestFindByID_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)

	_, err := svc.FindByID(context.Background(), 9999)
	require.ErrorIs(t, err, pathmapping.ErrNotFound)
}

func TestList(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, pathmapping.CreateInput{
		Host: "a", RemotePath: "/r1/", LocalPath: "/l1/",
	})
	require.NoError(t, err)
	_, err = svc.Create(ctx, pathmapping.CreateInput{
		Host: "b", RemotePath: "/r2/", LocalPath: "/l2/",
	})
	require.NoError(t, err)

	mappings, err := svc.List(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(mappings), 2)
}

func TestList_Empty(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)

	mappings, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, mappings)
}

func TestUpdate(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, pathmapping.CreateInput{
		Host: "old", RemotePath: "/old/remote/", LocalPath: "/old/local/",
	})
	require.NoError(t, err)

	newHost := "new-host"
	newRemote := "/new/remote"
	updated, err := svc.Update(ctx, created.ID, pathmapping.UpdateInput{
		Host:       &newHost,
		RemotePath: &newRemote,
	})
	require.NoError(t, err)
	assert.Equal(t, "new-host", updated.Host)
	assert.Equal(t, "/new/remote/", updated.RemotePath) // normalized
	assert.Equal(t, "/old/local/", updated.LocalPath)   // unchanged
}

func TestUpdate_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)

	_, err := svc.Update(context.Background(), 9999, pathmapping.UpdateInput{})
	require.ErrorIs(t, err, pathmapping.ErrNotFound)
}

func TestUpdate_ClearsRequiredField(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, pathmapping.CreateInput{
		Host: "h", RemotePath: "/r/", LocalPath: "/l/",
	})
	require.NoError(t, err)

	empty := ""
	_, err = svc.Update(ctx, created.ID, pathmapping.UpdateInput{RemotePath: &empty})
	require.ErrorIs(t, err, pathmapping.ErrInvalidInput)
}

func TestDelete(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, pathmapping.CreateInput{
		Host: "h", RemotePath: "/r/", LocalPath: "/l/",
	})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, created.ID)
	require.ErrorIs(t, err, pathmapping.ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)

	err := svc.Delete(context.Background(), 9999)
	require.ErrorIs(t, err, pathmapping.ErrNotFound)
}

func TestMapPath_Match(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, pathmapping.CreateInput{
		Host:       "sabnzbd",
		RemotePath: "/remote/downloads",
		LocalPath:  "/local/downloads",
	})
	require.NoError(t, err)

	got := svc.MapPath(ctx, "/remote/downloads/MyBook/book.epub")
	assert.Equal(t, "/local/downloads/MyBook/book.epub", got)
}

func TestMapPath_NoMatch(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, pathmapping.CreateInput{
		Host:       "host",
		RemotePath: "/remote/other/",
		LocalPath:  "/local/other/",
	})
	require.NoError(t, err)

	got := svc.MapPath(ctx, "/completely/different/path")
	assert.Equal(t, "/completely/different/path", got)
}

func TestMapPath_NoMappings(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	svc := pathmapping.New(db)

	got := svc.MapPath(context.Background(), "/any/path")
	assert.Equal(t, "/any/path", got)
}
