package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/woliveiras/bookaneer/internal/auth"
)

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			slog.Info("http",
				"method", req.Method,
				"path", req.URL.Path,
				"status", res.Status,
				"duration", time.Since(start).String(),
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
			)

			return nil
		}
	}
}

// Auth returns a middleware that validates API key or session.
func Auth(svc *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			// Check X-Api-Key header
			apiKey := c.Request().Header.Get("X-Api-Key")
			if apiKey == "" {
				// Check query parameter
				apiKey = c.QueryParam("apikey")
			}

			if apiKey != "" {
				// Validate against system API key
				if svc.ValidateAPIKey(ctx, apiKey) {
					// Store the validated API key
					c.Set("apiKey", apiKey)
					// Try to get user by API key
					user, err := svc.GetUserByAPIKey(ctx, apiKey)
					if err == nil {
						c.Set("user", user)
					}
					return next(c)
				}
				return echo.ErrUnauthorized
			}

			return echo.ErrUnauthorized
		}
	}
}
