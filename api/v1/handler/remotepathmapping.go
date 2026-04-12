package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/woliveiras/bookaneer/internal/core/pathmapping"
)

// RemotePathMappingHandler handles remote path mapping HTTP requests.
type RemotePathMappingHandler struct {
	svc *pathmapping.Service
}

// NewRemotePathMappingHandler creates a new handler.
func NewRemotePathMappingHandler(svc *pathmapping.Service) *RemotePathMappingHandler {
	return &RemotePathMappingHandler{svc: svc}
}

// Register registers the routes.
func (h *RemotePathMappingHandler) Register(g *echo.Group) {
	g.GET("/remotepathmapping", h.List)
	g.GET("/remotepathmapping/:id", h.GetByID)
	g.POST("/remotepathmapping", h.Create)
	g.PUT("/remotepathmapping/:id", h.Update)
	g.DELETE("/remotepathmapping/:id", h.Delete)
}

// List returns all remote path mappings.
func (h *RemotePathMappingHandler) List(c *echo.Context) error {
	mappings, err := h.svc.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list remote path mappings")
	}
	if mappings == nil {
		mappings = []pathmapping.RemotePathMapping{}
	}
	return c.JSON(http.StatusOK, mappings)
}

// GetByID returns a mapping by ID.
func (h *RemotePathMappingHandler) GetByID(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid mapping id")
	}

	m, err := h.svc.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pathmapping.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "mapping not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get mapping")
	}
	return c.JSON(http.StatusOK, m)
}

// Create creates a new mapping.
func (h *RemotePathMappingHandler) Create(c *echo.Context) error {
	var input pathmapping.CreateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	m, err := h.svc.Create(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, pathmapping.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "remote path and local path are required")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create mapping")
	}
	return c.JSON(http.StatusCreated, m)
}

// Update updates an existing mapping.
func (h *RemotePathMappingHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid mapping id")
	}

	var input pathmapping.UpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	m, err := h.svc.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, pathmapping.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "mapping not found")
		}
		if errors.Is(err, pathmapping.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "remote path and local path are required")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update mapping")
	}
	return c.JSON(http.StatusOK, m)
}

// Delete deletes a mapping.
func (h *RemotePathMappingHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid mapping id")
	}

	err = h.svc.Delete(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, pathmapping.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "mapping not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete mapping")
	}
	return c.NoContent(http.StatusNoContent)
}
