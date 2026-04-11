package qualityprofile_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/qualityprofile"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestNew(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	assert.NotNil(t, svc)
}

func TestCreate_Success(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	profile, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{
		Name:   "Standard",
		Cutoff: "epub",
		Items:  []qualityprofile.QualityItem{{Quality: "epub", Allowed: true}, {Quality: "mobi", Allowed: true}},
	})
	require.NoError(t, err)
	assert.NotZero(t, profile.ID)
	assert.Equal(t, "Standard", profile.Name)
	assert.Equal(t, "epub", profile.Cutoff)
	assert.Len(t, profile.Items, 2)
}

func TestCreate_EmptyName(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	_, err := svc.Create(context.Background(), qualityprofile.CreateQualityProfileInput{Cutoff: "epub"})
	require.ErrorIs(t, err, qualityprofile.ErrInvalidInput)
}

func TestFindByID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{
		Name: "Test", Cutoff: "epub", Items: []qualityprofile.QualityItem{{Quality: "epub", Allowed: true}},
	})
	require.NoError(t, err)

	found, err := svc.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test", found.Name)
	assert.Len(t, found.Items, 1)
}

func TestFindByID_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	_, err := svc.FindByID(context.Background(), 9999)
	require.ErrorIs(t, err, qualityprofile.ErrNotFound)
}

func TestList(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{Name: "P1", Cutoff: "epub"})
	require.NoError(t, err)
	_, err = svc.Create(ctx, qualityprofile.CreateQualityProfileInput{Name: "P2", Cutoff: "mobi"})
	require.NoError(t, err)

	profiles, err := svc.List(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(profiles), 2)
}

func TestUpdate_Name(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{Name: "Old", Cutoff: "epub"})
	require.NoError(t, err)

	newName := "New Name"
	updated, err := svc.Update(ctx, created.ID, qualityprofile.UpdateQualityProfileInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestUpdate_Cutoff(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{Name: "Test", Cutoff: "epub"})
	require.NoError(t, err)

	newCutoff := "mobi"
	updated, err := svc.Update(ctx, created.ID, qualityprofile.UpdateQualityProfileInput{Cutoff: &newCutoff})
	require.NoError(t, err)
	assert.Equal(t, "mobi", updated.Cutoff)
}

func TestUpdate_Items(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{
		Name: "Test", Cutoff: "epub", Items: []qualityprofile.QualityItem{{Quality: "epub", Allowed: true}},
	})
	require.NoError(t, err)

	newItems := []qualityprofile.QualityItem{{Quality: "mobi", Allowed: true}, {Quality: "pdf", Allowed: false}}
	updated, err := svc.Update(ctx, created.ID, qualityprofile.UpdateQualityProfileInput{Items: &newItems})
	require.NoError(t, err)
	assert.Len(t, updated.Items, 2)
}

func TestUpdate_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	name := "X"
	_, err := svc.Update(context.Background(), 9999, qualityprofile.UpdateQualityProfileInput{Name: &name})
	require.ErrorIs(t, err, qualityprofile.ErrNotFound)
}

func TestDelete(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{Name: "ToDelete", Cutoff: "epub"})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.FindByID(ctx, created.ID)
	require.ErrorIs(t, err, qualityprofile.ErrNotFound)
}

func TestDelete_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	err := svc.Delete(context.Background(), 9999)
	require.ErrorIs(t, err, qualityprofile.ErrNotFound)
}

func TestEnsureDefault(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	err := svc.EnsureDefault(ctx)
	require.NoError(t, err)

	profiles, err := svc.List(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, profiles)

	// Second call is idempotent
	err = svc.EnsureDefault(ctx)
	require.NoError(t, err)
}

func TestMarshalItems(t *testing.T) {
	items := []qualityprofile.QualityItem{
		{Quality: "epub", Allowed: true},
		{Quality: "mobi", Allowed: false},
	}
	data := qualityprofile.MarshalItems(items)
	assert.Contains(t, data, "epub")
	assert.Contains(t, data, "mobi")
}

func TestUnmarshalItems(t *testing.T) {
	data := `[{"quality":"epub","allowed":true},{"quality":"mobi","allowed":false}]`
	items := qualityprofile.UnmarshalItems(data)
	assert.Len(t, items, 2)
	assert.Equal(t, "epub", items[0].Quality)
	assert.True(t, items[0].Allowed)
	assert.Equal(t, "mobi", items[1].Quality)
	assert.False(t, items[1].Allowed)
}

func TestUnmarshalItems_Empty(t *testing.T) {
	items := qualityprofile.UnmarshalItems("")
	assert.Empty(t, items)
}

func TestUnmarshalItems_Invalid(t *testing.T) {
	items := qualityprofile.UnmarshalItems("not json")
	assert.Empty(t, items)
}

func TestDefaultProfile(t *testing.T) {
	p := qualityprofile.DefaultProfile()
	assert.NotNil(t, p)
	assert.NotEmpty(t, p.Name)
	assert.NotEmpty(t, p.Cutoff)
	assert.NotEmpty(t, p.Items)
}

func TestAllQualities(t *testing.T) {
	q := qualityprofile.AllQualities()
	assert.NotEmpty(t, q)
	assert.Contains(t, q, "epub")
}

func TestCreate_WithEmptyItems(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	profile, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{
		Name: "No Items",
	})
	require.NoError(t, err)
	assert.Equal(t, "epub", profile.Cutoff)
	assert.NotEmpty(t, profile.Items, "should use default items")
}

func TestDelete_InUse(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	profile, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{
		Name: "InUse", Cutoff: "epub",
	})
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `INSERT INTO root_folders (path, name, default_quality_profile_id) VALUES (?, ?, ?)`,
		"/tmp/test", "Test", profile.ID)
	require.NoError(t, err)

	err = svc.Delete(ctx, profile.ID)
	require.ErrorIs(t, err, qualityprofile.ErrInUse)
}

func TestUpdate_EmptyInput(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{
		Name: "Unchanged", Cutoff: "epub",
	})
	require.NoError(t, err)

	updated, err := svc.Update(ctx, created.ID, qualityprofile.UpdateQualityProfileInput{})
	require.NoError(t, err)
	assert.Equal(t, "Unchanged", updated.Name)
}

func TestList_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)

	profiles, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, profiles)
}

func TestFindByID_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	_ = db.Close()

	_, err := svc.FindByID(context.Background(), 1)
	require.Error(t, err)
	require.NotErrorIs(t, err, qualityprofile.ErrNotFound)
}

func TestList_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	_ = db.Close()

	_, err := svc.List(context.Background())
	require.Error(t, err)
}

func TestCreate_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	_ = db.Close()

	_, err := svc.Create(context.Background(), qualityprofile.CreateQualityProfileInput{Name: "Test"})
	require.Error(t, err)
}

func TestDelete_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	db.Close()

	err := svc.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NotErrorIs(t, err, qualityprofile.ErrNotFound)
	require.NotErrorIs(t, err, qualityprofile.ErrInUse)
}

func TestEnsureDefault_DBClosed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	db.Close()

	err := svc.EnsureDefault(context.Background())
	require.Error(t, err)
}

func TestEnsureDefault_WhenDefaultExists(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := qualityprofile.New(db)
	ctx := context.Background()

	_, err := svc.Create(ctx, qualityprofile.CreateQualityProfileInput{Name: "Existing", Cutoff: "epub"})
	require.NoError(t, err)

	err = svc.EnsureDefault(ctx)
	require.NoError(t, err)

	profiles, err := svc.List(ctx)
	require.NoError(t, err)
	assert.Len(t, profiles, 1, "should not create additional default when one exists")
}
