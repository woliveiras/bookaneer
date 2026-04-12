package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/woliveiras/bookaneer/internal/notification"
)

// NotificationHandler handles notification CRUD and test endpoints.
type NotificationHandler struct {
	svc *notification.Service
}

// NewNotificationHandler creates a new notification handler.
func NewNotificationHandler(svc *notification.Service) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// Register registers notification routes on the given group.
func (h *NotificationHandler) Register(g *echo.Group) {
	g.GET("/notification", h.List)
	g.GET("/notification/:id", h.Get)
	g.POST("/notification", h.Create)
	g.PUT("/notification/:id", h.Update)
	g.DELETE("/notification/:id", h.Delete)
	g.POST("/notification/:id/test", h.Test)
}

// List returns all notification configs.
func (h *NotificationHandler) List(c *echo.Context) error {
	configs, err := h.svc.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if configs == nil {
		configs = []notification.Config{}
	}
	return c.JSON(http.StatusOK, configs)
}

// Get returns a single notification config.
func (h *NotificationHandler) Get(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cfg, err := h.svc.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "notification not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, cfg)
}

// Create creates a new notification config.
func (h *NotificationHandler) Create(c *echo.Context) error {
	var input notification.CreateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if input.Name == "" || input.Type == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and type are required")
	}
	cfg, err := h.svc.Create(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, cfg)
}

// Update modifies an existing notification config.
func (h *NotificationHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var input notification.UpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	cfg, err := h.svc.Update(c.Request().Context(), id, input)
	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "notification not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, cfg)
}

// Delete removes a notification config.
func (h *NotificationHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	err = h.svc.Delete(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "notification not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// Test sends a test notification to the specified channel.
func (h *NotificationHandler) Test(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.TestChannel(c.Request().Context(), id); err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "notification not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
