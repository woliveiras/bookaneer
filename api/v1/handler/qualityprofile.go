package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/woliveiras/bookaneer/internal/core/qualityprofile"
)

// QualityProfileHandler handles quality profile-related HTTP requests.
type QualityProfileHandler struct {
	svc *qualityprofile.Service
}

// NewQualityProfileHandler creates a new quality profile handler.
func NewQualityProfileHandler(svc *qualityprofile.Service) *QualityProfileHandler {
	return &QualityProfileHandler{svc: svc}
}

// Register registers the quality profile routes.
func (h *QualityProfileHandler) Register(g *echo.Group) {
	g.GET("/qualityprofile", h.List)
	g.GET("/qualityprofile/:id", h.GetByID)
	g.POST("/qualityprofile", h.Create)
	g.PUT("/qualityprofile/:id", h.Update)
	g.DELETE("/qualityprofile/:id", h.Delete)
}

// List returns all quality profiles.
func (h *QualityProfileHandler) List(c echo.Context) error {
	profiles, err := h.svc.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list quality profiles")
	}

	return c.JSON(http.StatusOK, profiles)
}

// GetByID returns a quality profile by ID.
func (h *QualityProfileHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid quality profile id")
	}

	qp, err := h.svc.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, qualityprofile.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "quality profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get quality profile")
	}

	return c.JSON(http.StatusOK, qp)
}

// Create creates a new quality profile.
func (h *QualityProfileHandler) Create(c echo.Context) error {
	var input qualityprofile.CreateQualityProfileInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	qp, err := h.svc.Create(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, qualityprofile.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid quality profile data")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create quality profile")
	}

	return c.JSON(http.StatusCreated, qp)
}

// Update updates an existing quality profile.
func (h *QualityProfileHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid quality profile id")
	}

	var input qualityprofile.UpdateQualityProfileInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	qp, err := h.svc.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, qualityprofile.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "quality profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update quality profile")
	}

	return c.JSON(http.StatusOK, qp)
}

// Delete deletes a quality profile.
func (h *QualityProfileHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid quality profile id")
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, qualityprofile.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "quality profile not found")
		}
		if errors.Is(err, qualityprofile.ErrInUse) {
			return echo.NewHTTPError(http.StatusConflict, "quality profile is in use")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete quality profile")
	}

	return c.NoContent(http.StatusNoContent)
}
