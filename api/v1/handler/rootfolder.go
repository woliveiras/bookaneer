package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/woliveiras/bookaneer/internal/core/rootfolder"
)

// RootFolderHandler handles root folder-related HTTP requests.
type RootFolderHandler struct {
	svc *rootfolder.Service
}

// NewRootFolderHandler creates a new root folder handler.
func NewRootFolderHandler(svc *rootfolder.Service) *RootFolderHandler {
	return &RootFolderHandler{svc: svc}
}

// Register registers the root folder routes.
func (h *RootFolderHandler) Register(g *echo.Group) {
	g.GET("/rootfolder", h.List)
	g.GET("/rootfolder/:id", h.GetByID)
	g.POST("/rootfolder", h.Create)
	g.PUT("/rootfolder/:id", h.Update)
	g.DELETE("/rootfolder/:id", h.Delete)
	g.POST("/rootfolder/:id/migrate", h.Migrate)
}

// List returns all root folders.
func (h *RootFolderHandler) List(c *echo.Context) error {
	folders, err := h.svc.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list root folders")
	}

	return c.JSON(http.StatusOK, folders)
}

// GetByID returns a root folder by ID.
func (h *RootFolderHandler) GetByID(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid root folder id")
	}

	rf, err := h.svc.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, rootfolder.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "root folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get root folder")
	}

	return c.JSON(http.StatusOK, rf)
}

// Create creates a new root folder.
func (h *RootFolderHandler) Create(c *echo.Context) error {
	var input rootfolder.CreateRootFolderInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	rf, err := h.svc.Create(c.Request().Context(), input)
	if err != nil {
		if errors.Is(err, rootfolder.ErrInvalidInput) {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid root folder data")
		}
		if errors.Is(err, rootfolder.ErrPathNotAccessible) {
			return echo.NewHTTPError(http.StatusBadRequest, "path is not accessible")
		}
		if errors.Is(err, rootfolder.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "root folder already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create root folder")
	}

	return c.JSON(http.StatusCreated, rf)
}

// Update updates an existing root folder.
func (h *RootFolderHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid root folder id")
	}

	var input rootfolder.UpdateRootFolderInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	rf, err := h.svc.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, rootfolder.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "root folder not found")
		}
		if errors.Is(err, rootfolder.ErrPathNotAccessible) {
			return echo.NewHTTPError(http.StatusBadRequest, "path is not accessible")
		}
		if errors.Is(err, rootfolder.ErrDuplicate) {
			return echo.NewHTTPError(http.StatusConflict, "root folder already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update root folder")
	}

	return c.JSON(http.StatusOK, rf)
}

// Delete deletes a root folder.
func (h *RootFolderHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid root folder id")
	}

	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, rootfolder.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "root folder not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete root folder")
	}

	return c.NoContent(http.StatusNoContent)
}

// MigrateInput is the input for the migrate endpoint.
type MigrateInput struct {
	NewPath string `json:"newPath"`
}

// Migrate moves all files from the current root folder path to a new path.
// This is a blocking operation that moves all files and updates database paths.
func (h *RootFolderHandler) Migrate(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid root folder id")
	}

	var input MigrateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if input.NewPath == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "newPath is required")
	}

	rf, err := h.svc.MoveRootFolder(c.Request().Context(), id, input.NewPath)
	if err != nil {
		if errors.Is(err, rootfolder.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "root folder not found")
		}
		if errors.Is(err, rootfolder.ErrPathNotAccessible) {
			return echo.NewHTTPError(http.StatusBadRequest, "path is not accessible")
		}
		// Include detailed error for migration failures
		return echo.NewHTTPError(http.StatusInternalServerError, "migration failed: "+err.Error())
	}

	return c.JSON(http.StatusOK, rf)
}
