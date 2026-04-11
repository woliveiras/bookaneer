package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/woliveiras/bookaneer/internal/auth"
	"github.com/woliveiras/bookaneer/internal/config"
)

// SettingsHandler handles settings endpoints.
type SettingsHandler struct {
	authSvc *auth.Service
	cfg     *config.Config
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(authSvc *auth.Service, cfg *config.Config) *SettingsHandler {
	return &SettingsHandler{
		authSvc: authSvc,
		cfg:     cfg,
	}
}

// GeneralSettingsResponse is returned by the general settings endpoint.
type GeneralSettingsResponse struct {
	APIKey                string                        `json:"apiKey"`
	BindAddress           string                        `json:"bindAddress"`
	Port                  int                           `json:"port"`
	DataDir               string                        `json:"dataDir"`
	LibraryDir            string                        `json:"libraryDir"`
	LogLevel              string                        `json:"logLevel"`
	CustomProvidersEnable bool                          `json:"customProvidersEnabled"`
	CustomProvidersActive []config.CustomProviderConfig `json:"customProvidersActive"`
}

// GetGeneral returns general application settings including the API key.
// GET /api/v1/settings/general
func (h *SettingsHandler) GetGeneral(c echo.Context) error {
	ctx := c.Request().Context()

	apiKey, err := h.authSvc.GetAPIKey(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get API key")
	}

	return c.JSON(http.StatusOK, GeneralSettingsResponse{
		APIKey:                apiKey,
		BindAddress:           h.cfg.BindAddress,
		Port:                  h.cfg.Port,
		DataDir:               h.cfg.DataDir,
		LibraryDir:            h.cfg.LibraryDir,
		LogLevel:              h.cfg.LogLevel,
		CustomProvidersEnable: h.cfg.CustomProvidersEnable,
		CustomProvidersActive: activeCustomProviders(h.cfg),
	})
}

func activeCustomProviders(cfg *config.Config) []config.CustomProviderConfig {
	if cfg == nil || !cfg.CustomProvidersEnable || len(cfg.CustomProviders) == 0 {
		return []config.CustomProviderConfig{}
	}

	active := make([]config.CustomProviderConfig, 0, len(cfg.CustomProviders))
	for _, cp := range cfg.CustomProviders {
		if cp.Name == "" || cp.Domain == "" {
			continue
		}
		active = append(active, cp)
	}

	return active
}

// Register registers the settings routes.
func (h *SettingsHandler) Register(g *echo.Group) {
	g.GET("/settings/general", h.GetGeneral)
}
