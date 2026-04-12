package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/woliveiras/bookaneer/internal/auth"
)

// Logger returns a middleware that logs HTTP requests.
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			req := c.Request()
			res, _ := echo.UnwrapResponse(c.Response())

			status := 0
			if res != nil {
				status = res.Status
			}

			slog.Info("http",
				"method", req.Method,
				"path", req.URL.Path,
				"status", status,
				"duration", time.Since(start).String(),
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
			)

			return err
		}
	}
}

// Auth returns a middleware that validates API key or session.
func Auth(svc *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()

			// Check X-Api-Key header
			apiKey := c.Request().Header.Get("X-Api-Key")
			if apiKey == "" {
				// Check query parameter (support both "apikey" and "key")
				apiKey = c.QueryParam("apikey")
			}
			if apiKey == "" {
				apiKey = c.QueryParam("key")
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
