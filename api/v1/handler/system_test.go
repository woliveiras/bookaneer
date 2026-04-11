package handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/config"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func setupSystemHandler(t *testing.T) (*SystemHandler, *echo.Echo) {
	t.Helper()
	db := testutil.OpenTestDB(t)

	tmpDir := t.TempDir()
	cfg := &config.Config{
		DataDir:    tmpDir,
		LibraryDir: tmpDir,
	}

	h := NewSystemHandler("test", "now", cfg, db)
	e := echo.New()
	return h, e
}

func TestStatus(t *testing.T) {
	h, e := setupSystemHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/system/status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Status(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp StatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "test", resp.Version)
	assert.Equal(t, "now", resp.BuildTime)
	assert.NotEmpty(t, resp.OSName)
	assert.NotEmpty(t, resp.RuntimeVersion)
}

func TestHealth(t *testing.T) {
	h, e := setupSystemHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/system/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Health(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp HealthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Status)
	assert.NotEmpty(t, resp.Checks)

	// Should have database, dataDir, libraryDir, runtime checks
	checkNames := make(map[string]bool)
	for _, ch := range resp.Checks {
		checkNames[ch.Name] = true
	}
	assert.True(t, checkNames["database"])
	assert.True(t, checkNames["runtime"])
}

func TestBackup(t *testing.T) {
	h, e := setupSystemHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/system/backup", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Backup(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "bookaneer-")
	assert.Contains(t, rec.Header().Get("Content-Disposition"), ".zip")

	// Verify it's a valid zip with bookaneer.db
	body := rec.Body.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	require.NoError(t, err)

	var foundDB bool
	for _, f := range zr.File {
		if f.Name == "bookaneer.db" {
			foundDB = true
			rc, err := f.Open()
			require.NoError(t, err)
			data, _ := io.ReadAll(rc)
			_ = rc.Close()
			assert.NotEmpty(t, data)
		}
	}
	assert.True(t, foundDB, "zip should contain bookaneer.db")
}

func TestRestore_MissingFile(t *testing.T) {
	h, e := setupSystemHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/system/restore", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Restore(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestRestore_ValidZip(t *testing.T) {
	h, e := setupSystemHandler(t)

	// Create a valid backup zip in memory
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("bookaneer.db")
	require.NoError(t, err)
	_, _ = w.Write([]byte("SQLite format 3\x00")) // Fake DB header
	_ = zw.Close()

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("backup", "backup.zip")
	require.NoError(t, err)
	_, _ = part.Write(buf.Bytes())
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/system/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.Restore(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])

	// Verify the restore file was created
	restoredPath := filepath.Join(h.cfg.DataDir, "bookaneer-restore.db")
	_, err = os.Stat(restoredPath)
	assert.NoError(t, err)
}

func TestRestore_InvalidZip(t *testing.T) {
	h, e := setupSystemHandler(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("backup", "notazip.txt")
	require.NoError(t, err)
	_, _ = part.Write([]byte("this is not a zip file"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/system/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.Restore(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestRestore_ZipWithoutDB(t *testing.T) {
	h, e := setupSystemHandler(t)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("other.txt")
	require.NoError(t, err)
	_, _ = w.Write([]byte("hello"))
	_ = zw.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("backup", "backup.zip")
	require.NoError(t, err)
	_, _ = part.Write(buf.Bytes())
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/system/restore", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.Restore(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestCheckDiskSpace(t *testing.T) {
	t.Run("existing dir", func(t *testing.T) {
		check := checkDiskSpace("test", os.TempDir())
		assert.Equal(t, "ok", check.Status)
	})

	t.Run("missing dir", func(t *testing.T) {
		check := checkDiskSpace("test", "/nonexistent/path/xyz")
		assert.Equal(t, "degraded", check.Status)
	})
}
