package handler

import (
	"net/http"
	"os"
	"runtime"
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
}

// NewSystemHandler creates a new system handler.
func NewSystemHandler(version, buildTime string, cfg *config.Config) *SystemHandler {
	return &SystemHandler{
		version:   version,
		buildTime: buildTime,
		cfg:       cfg,
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

// HealthResponse is returned by the health endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// Health returns a simple health check response.
func (h *SystemHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
