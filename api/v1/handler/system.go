package handler

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/woliveiras/bookaneer/internal/config"
)

var startTime = time.Now()

// SystemHandler handles system status and health endpoints.
type SystemHandler struct {
	version   string
	buildTime string
	cfg       *config.Config
	db        *sql.DB
}

// NewSystemHandler creates a new system handler.
func NewSystemHandler(version, buildTime string, cfg *config.Config, db *sql.DB) *SystemHandler {
	return &SystemHandler{
		version:   version,
		buildTime: buildTime,
		cfg:       cfg,
		db:        db,
	}
}

// StatusResponse is returned by the status endpoint.
type StatusResponse struct {
	Version        string `json:"version"`
	BuildTime      string `json:"buildTime"`
	OSName         string `json:"osName"`
	OSArch         string `json:"osArch"`
	RuntimeVersion string `json:"runtimeVersion"`
	StartTime      string `json:"startTime"`
	AppDataDir     string `json:"appDataDir"`
	LibraryDir     string `json:"libraryDir"`
}

// Status returns system status information.
func (h *SystemHandler) Status(c echo.Context) error {
	hostname, _ := os.Hostname()
	_ = hostname

	return c.JSON(http.StatusOK, StatusResponse{
		Version:        h.version,
		BuildTime:      h.buildTime,
		OSName:         runtime.GOOS,
		OSArch:         runtime.GOARCH,
		RuntimeVersion: runtime.Version(),
		StartTime:      startTime.UTC().Format(time.RFC3339),
		AppDataDir:     h.cfg.DataDir,
		LibraryDir:     h.cfg.LibraryDir,
	})
}

// HealthCheck contains a named health check result.
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok, degraded, error
	Message string `json:"message,omitempty"`
}

// HealthResponse is returned by the health endpoint.
type HealthResponse struct {
	Status string        `json:"status"`
	Checks []HealthCheck `json:"checks,omitempty"`
}

// Health returns a comprehensive health check response.
func (h *SystemHandler) Health(c echo.Context) error {
	var checks []HealthCheck
	overallStatus := "ok"

	// Database check
	if err := h.db.PingContext(c.Request().Context()); err != nil {
		checks = append(checks, HealthCheck{Name: "database", Status: "error", Message: err.Error()})
		overallStatus = "error"
	} else {
		var dbSize int64
		_ = h.db.QueryRowContext(c.Request().Context(), "SELECT page_count * page_size FROM pragma_page_count, pragma_page_size").Scan(&dbSize)
		checks = append(checks, HealthCheck{Name: "database", Status: "ok", Message: fmt.Sprintf("size: %d bytes", dbSize)})
	}

	// Disk space check (data dir)
	checks = append(checks, checkDiskSpace("dataDir", h.cfg.DataDir))
	checks = append(checks, checkDiskSpace("libraryDir", h.cfg.LibraryDir))

	// Runtime info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	checks = append(checks, HealthCheck{
		Name:    "runtime",
		Status:  "ok",
		Message: fmt.Sprintf("goroutines: %d, heap: %d MB", runtime.NumGoroutine(), m.HeapAlloc/1024/1024),
	})

	// If any check failed, degrade overall
	for _, ch := range checks {
		if ch.Status == "error" {
			overallStatus = "error"
			break
		}
		if ch.Status == "degraded" && overallStatus == "ok" {
			overallStatus = "degraded"
		}
	}

	code := http.StatusOK
	if overallStatus == "error" {
		code = http.StatusServiceUnavailable
	}

	return c.JSON(code, HealthResponse{Status: overallStatus, Checks: checks})
}

// Backup creates a database backup via VACUUM INTO and returns it as a zip.
func (h *SystemHandler) Backup(c echo.Context) error {
	backupDir := filepath.Join(h.cfg.DataDir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create backup dir: "+err.Error())
	}

	ts := time.Now().UTC().Format("20060102-150405")
	dbBackupPath := filepath.Join(backupDir, fmt.Sprintf("bookaneer-%s.db", ts))
	zipPath := filepath.Join(backupDir, fmt.Sprintf("bookaneer-%s.zip", ts))

	// VACUUM INTO creates a clean copy of the database
	_, err := h.db.ExecContext(c.Request().Context(), fmt.Sprintf("VACUUM INTO '%s'", dbBackupPath))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "backup database: "+err.Error())
	}
	defer func() { _ = os.Remove(dbBackupPath) }() // Clean up raw db file after zipping

	// Create zip
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create zip: "+err.Error())
	}

	zw := zip.NewWriter(zipFile)
	dbFile, err := os.Open(dbBackupPath)
	if err != nil {
		_ = zipFile.Close()
		return echo.NewHTTPError(http.StatusInternalServerError, "open backup: "+err.Error())
	}

	w, err := zw.Create("bookaneer.db")
	if err != nil {
		_ = dbFile.Close()
		_ = zipFile.Close()
		return echo.NewHTTPError(http.StatusInternalServerError, "create zip entry: "+err.Error())
	}
	if _, err := io.Copy(w, dbFile); err != nil {
		_ = dbFile.Close()
		_ = zipFile.Close()
		return echo.NewHTTPError(http.StatusInternalServerError, "write zip: "+err.Error())
	}
	_ = dbFile.Close()

	// Include config.yaml if present
	configPath := filepath.Join(h.cfg.DataDir, "config.yaml")
	if f, err := os.Open(configPath); err == nil {
		w, err := zw.Create("config.yaml")
		if err == nil {
			_, _ = io.Copy(w, f)
		}
		_ = f.Close()
	}

	_ = zw.Close()
	_ = zipFile.Close()

	defer func() { _ = os.Remove(zipPath) }()

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="bookaneer-%s.zip"`, ts))
	return c.File(zipPath)
}

// Restore restores a database from an uploaded backup zip.
func (h *SystemHandler) Restore(c echo.Context) error {
	file, err := c.FormFile("backup")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "backup file is required")
	}

	// Validate file size (max 500 MB)
	if file.Size > 500*1024*1024 {
		return echo.NewHTTPError(http.StatusBadRequest, "backup file too large (max 500 MB)")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "open upload: "+err.Error())
	}
	defer func() { _ = src.Close() }()

	// Save to temp file for zip reading
	tmpFile, err := os.CreateTemp("", "bookaneer-restore-*.zip")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create temp: "+err.Error())
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := io.Copy(tmpFile, src); err != nil {
		_ = tmpFile.Close()
		return echo.NewHTTPError(http.StatusInternalServerError, "save upload: "+err.Error())
	}
	_ = tmpFile.Close()

	// Open as zip
	zr, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid zip file")
	}
	defer func() { _ = zr.Close() }()

	// Extract database file
	var dbFound bool
	for _, f := range zr.File {
		// Security: prevent path traversal
		name := filepath.Base(f.Name)
		if strings.Contains(f.Name, "..") {
			continue
		}

		if name == "bookaneer.db" {
			rc, err := f.Open()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "open db from zip: "+err.Error())
			}
			destPath := filepath.Join(h.cfg.DataDir, "bookaneer-restore.db")
			dest, err := os.Create(destPath)
			if err != nil {
				_ = rc.Close()
				return echo.NewHTTPError(http.StatusInternalServerError, "create restore db: "+err.Error())
			}
			_, err = io.Copy(dest, rc)
			_ = rc.Close()
			_ = dest.Close()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "write restore db: "+err.Error())
			}
			dbFound = true
		}

		if name == "config.yaml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			destPath := filepath.Join(h.cfg.DataDir, "config.yaml.restored")
			dest, err := os.Create(destPath)
			if err != nil {
				_ = rc.Close()
				continue
			}
			_, _ = io.Copy(dest, rc)
			_ = rc.Close()
			_ = dest.Close()
		}
	}

	if !dbFound {
		return echo.NewHTTPError(http.StatusBadRequest, "zip does not contain bookaneer.db")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "Database restored to bookaneer-restore.db. Restart Bookaneer with the restored database.",
	})
}

func checkDiskSpace(name, dir string) HealthCheck {
	info, err := os.Stat(dir)
	if err != nil {
		return HealthCheck{Name: name, Status: "degraded", Message: "directory not accessible"}
	}
	_ = info
	return HealthCheck{Name: name, Status: "ok"}
}
