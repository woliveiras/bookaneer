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
	user, ok := c.Get("user").(*auth.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	return c.JSON(http.StatusOK, user)
}

// Logout is a no-op for API key auth (client just removes the key).
func (h *AuthHandler) Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
