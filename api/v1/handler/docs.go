package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// DocsHandler serves the API documentation page.
type DocsHandler struct{}

// NewDocsHandler creates a new docs handler.
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// Register registers docs routes on the given group.
func (h *DocsHandler) Register(g *echo.Group) {
	g.GET("/docs", h.SwaggerUI)
	g.GET("/docs/openapi.json", h.OpenAPISpec)
}

// SwaggerUI serves a minimal Swagger UI HTML page that loads the OpenAPI spec.
func (h *DocsHandler) SwaggerUI(c *echo.Context) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Bookaneer API Docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({ url: "/api/v1/docs/openapi.json", dom_id: "#swagger-ui" });
  </script>
</body>
</html>`
	return c.HTML(http.StatusOK, html)
}

// OpenAPISpec returns the OpenAPI 3.0 specification.
func (h *DocsHandler) OpenAPISpec(c *echo.Context) error {
	spec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Bookaneer API",
			"version":     "1.0.0",
			"description": "Self-hosted ebook collection manager API",
		},
		"servers": []map[string]string{
			{"url": "/api/v1", "description": "Default"},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"ApiKey": map[string]interface{}{
					"type": "apiKey",
					"in":   "header",
					"name": "X-Api-Key",
				},
			},
		},
		"security": []map[string][]string{
			{"ApiKey": {}},
		},
		"paths": map[string]interface{}{
			"/system/status": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "System status",
					"tags":    []string{"System"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "System status info"},
					},
				},
			},
			"/system/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Health check",
					"tags":    []string{"System"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "Health check with component status"},
					},
				},
			},
			"/system/backup": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Download database backup",
					"tags":    []string{"System"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "ZIP file containing database backup"},
					},
				},
			},
			"/system/restore": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Restore database from backup",
					"tags":    []string{"System"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "Restore confirmation"},
					},
				},
			},
			"/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"summary": "Login",
					"tags":    []string{"Auth"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "Login successful"},
					},
				},
			},
			"/author": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List authors",
					"tags":    []string{"Authors"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "List of authors"},
					},
				},
				"post": map[string]interface{}{
					"summary": "Create author",
					"tags":    []string{"Authors"},
					"responses": map[string]interface{}{
						"201": map[string]string{"description": "Author created"},
					},
				},
			},
			"/book": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List books",
					"tags":    []string{"Books"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "List of books"},
					},
				},
				"post": map[string]interface{}{
					"summary": "Create book",
					"tags":    []string{"Books"},
					"responses": map[string]interface{}{
						"201": map[string]string{"description": "Book created"},
					},
				},
			},
			"/search": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Search indexers",
					"tags":    []string{"Search"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "Search results"},
					},
				},
			},
			"/download": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List download clients",
					"tags":    []string{"Download"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "List of download clients"},
					},
				},
			},
			"/notification": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List notification channels",
					"tags":    []string{"Notifications"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "List of notification channels"},
					},
				},
				"post": map[string]interface{}{
					"summary": "Create notification channel",
					"tags":    []string{"Notifications"},
					"responses": map[string]interface{}{
						"201": map[string]string{"description": "Notification channel created"},
					},
				},
			},
			"/wanted": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List wanted books",
					"tags":    []string{"Wanted"},
					"responses": map[string]interface{}{
						"200": map[string]string{"description": "List of wanted books"},
					},
				},
			},
			"/ws": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "WebSocket connection for real-time events",
					"tags":    []string{"WebSocket"},
					"responses": map[string]interface{}{
						"101": map[string]string{"description": "WebSocket upgrade"},
					},
				},
			},
		},
	}
	return c.JSON(http.StatusOK, spec)
}
