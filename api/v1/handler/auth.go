package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/woliveiras/bookaneer/internal/auth"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	svc *auth.Service
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// LoginRequest is the request body for login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	User   *auth.User `json:"user"`
	APIKey string     `json:"apiKey"`
}

// Login authenticates a user and returns their API key.
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "username and password required")
	}

	user, err := h.svc.Authenticate(c.Request().Context(), req.Username, req.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "authentication failed")
	}

	return c.JSON(http.StatusOK, LoginResponse{
		User:   user,
		APIKey: user.APIKey,
	})
}

// Me returns the current authenticated user.
func (h *AuthHandler) Me(c echo.Context) error {
	// Check if there's a user (from user-specific API key)
	if user, ok := c.Get("user").(*auth.User); ok {
		return c.JSON(http.StatusOK, user)
	}

	// Check if authenticated with system API key
	if apiKey, ok := c.Get("apiKey").(string); ok && apiKey != "" {
		// Return a synthetic admin user for system API key
		return c.JSON(http.StatusOK, &auth.User{
			ID:        0,
			Username:  "admin",
			Role:      "admin",
			APIKey:    apiKey,
			CreatedAt: "system",
		})
	}

	return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
}

// Logout is a no-op for API key auth (client just removes the key).
func (h *AuthHandler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
