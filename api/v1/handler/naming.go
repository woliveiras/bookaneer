package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/woliveiras/bookaneer/internal/core/naming"
)

// NamingHandler handles naming settings and preview endpoints.
type NamingHandler struct {
	engine *naming.Engine
}

// NewNamingHandler creates a new naming handler.
func NewNamingHandler(engine *naming.Engine) *NamingHandler {
	return &NamingHandler{engine: engine}
}

// NamingSettingsResponse is the response for naming settings.
type NamingSettingsResponse struct {
	Enabled            bool   `json:"enabled"`
	AuthorFolderFormat string `json:"authorFolderFormat"`
	BookFileFormat     string `json:"bookFileFormat"`
	ReplaceSpaces      bool   `json:"replaceSpaces"`
	ColonReplacement   string `json:"colonReplacement"`
}

// NamingSettingsRequest is the request for updating naming settings.
type NamingSettingsRequest struct {
	Enabled            *bool   `json:"enabled"`
	AuthorFolderFormat *string `json:"authorFolderFormat"`
	BookFileFormat     *string `json:"bookFileFormat"`
	ReplaceSpaces      *bool   `json:"replaceSpaces"`
	ColonReplacement   *string `json:"colonReplacement"`
}

// NamingPreviewRequest is the request for previewing a naming template.
type NamingPreviewRequest struct {
	AuthorFolderFormat string `json:"authorFolderFormat"`
	BookFileFormat     string `json:"bookFileFormat"`
	ReplaceSpaces      bool   `json:"replaceSpaces"`
	ColonReplacement   string `json:"colonReplacement"`
}

// NamingPreviewResponse is the response for a naming preview.
type NamingPreviewResponse struct {
	AuthorFolder string `json:"authorFolder"`
	Filename     string `json:"filename"`
	RelativePath string `json:"relativePath"`
	FullPath     string `json:"fullPath"`
}

// GetSettings returns the current naming settings.
// GET /api/v1/naming
func (h *NamingHandler) GetSettings(c echo.Context) error {
	ctx := c.Request().Context()

	s, err := h.engine.LoadSettings(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load naming settings")
	}

	return c.JSON(http.StatusOK, NamingSettingsResponse{
		Enabled:            s.Enabled,
		AuthorFolderFormat: s.AuthorFolderFormat,
		BookFileFormat:     s.BookFileFormat,
		ReplaceSpaces:      s.ReplaceSpaces,
		ColonReplacement:   s.ColonReplacement,
	})
}

// UpdateSettings updates naming settings.
// PUT /api/v1/naming
func (h *NamingHandler) UpdateSettings(c echo.Context) error {
	ctx := c.Request().Context()

	var req NamingSettingsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Load current settings as base
	s, err := h.engine.LoadSettings(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load naming settings")
	}

	// Apply partial updates
	if req.Enabled != nil {
		s.Enabled = *req.Enabled
	}
	if req.AuthorFolderFormat != nil {
		if !naming.IsValidTemplate(*req.AuthorFolderFormat) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid author folder template")
		}
		s.AuthorFolderFormat = *req.AuthorFolderFormat
	}
	if req.BookFileFormat != nil {
		if !naming.IsValidTemplate(*req.BookFileFormat) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid book file template")
		}
		s.BookFileFormat = *req.BookFileFormat
	}
	if req.ReplaceSpaces != nil {
		s.ReplaceSpaces = *req.ReplaceSpaces
	}
	if req.ColonReplacement != nil {
		valid := map[string]bool{"dash": true, "space": true, "delete": true}
		if !valid[*req.ColonReplacement] {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid colon replacement: must be dash, space, or delete")
		}
		s.ColonReplacement = *req.ColonReplacement
	}

	if err := h.engine.SaveSettings(ctx, s); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save naming settings")
	}

	return c.JSON(http.StatusOK, NamingSettingsResponse{
		Enabled:            s.Enabled,
		AuthorFolderFormat: s.AuthorFolderFormat,
		BookFileFormat:     s.BookFileFormat,
		ReplaceSpaces:      s.ReplaceSpaces,
		ColonReplacement:   s.ColonReplacement,
	})
}

// Preview returns a preview of how a book file would be named.
// POST /api/v1/naming/preview
func (h *NamingHandler) Preview(c echo.Context) error {
	var req NamingPreviewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	s := &naming.Settings{
		Enabled:            true,
		AuthorFolderFormat: req.AuthorFolderFormat,
		BookFileFormat:     req.BookFileFormat,
		ReplaceSpaces:      req.ReplaceSpaces,
		ColonReplacement:   req.ColonReplacement,
	}

	if s.AuthorFolderFormat == "" {
		s.AuthorFolderFormat = "$Author"
	}
	if s.BookFileFormat == "" {
		s.BookFileFormat = "$Author - $Title"
	}
	if s.ColonReplacement == "" {
		s.ColonReplacement = "dash"
	}

	sample := naming.SampleContext()
	result := h.engine.Preview("/books", sample, s)

	return c.JSON(http.StatusOK, NamingPreviewResponse{
		AuthorFolder: result.AuthorFolder,
		Filename:     result.Filename,
		RelativePath: result.RelativePath,
		FullPath:     result.FullPath,
	})
}

// Register registers the naming routes.
func (h *NamingHandler) Register(g *echo.Group) {
	g.GET("/naming", h.GetSettings)
	g.PUT("/naming", h.UpdateSettings)
	g.POST("/naming/preview", h.Preview)
	g.POST("/naming/rename/preview", h.PreviewRenameAll)
	g.POST("/naming/rename", h.RenameAll)
}

// PreviewRenameAll previews what a batch rename would do.
// POST /api/v1/naming/rename/preview
func (h *NamingHandler) PreviewRenameAll(c echo.Context) error {
	ctx := c.Request().Context()

	result, err := h.engine.PreviewRenameAll(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to preview rename")
	}

	return c.JSON(http.StatusOK, result)
}

// RenameAll renames all library files according to naming settings.
// POST /api/v1/naming/rename
func (h *NamingHandler) RenameAll(c echo.Context) error {
	ctx := c.Request().Context()

	result, err := h.engine.RenameAll(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to rename files")
	}

	return c.JSON(http.StatusOK, result)
}
