package direct

import (
	"mime"
	"net/http"
	"strings"
)

// extractFilename extracts filename from Content-Disposition header or URL.
// If a meaningful fallback name is provided (contains ebook extension), use it directly.
// This ensures books are saved with semantic names like "Author - Title.epub"
// instead of server-side IDs like "bv00180a.epub".
func (c *Client) extractFilename(resp *http.Response, fallback string) string {
	// Check if fallback already has a valid ebook extension - use it directly
	lowerFallback := strings.ToLower(fallback)
	validExtensions := []string{".epub", ".pdf", ".mobi", ".azw", ".azw3", ".fb2", ".cbz", ".cbr"}
	for _, ext := range validExtensions {
		if strings.HasSuffix(lowerFallback, ext) {
			return sanitizeFilename(fallback)
		}
	}

	// Fallback doesn't have extension, try to get extension from response
	ext := ""

	// Try Content-Disposition header
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil && params["filename"] != "" {
			serverFilename := params["filename"]
			if idx := strings.LastIndex(serverFilename, "."); idx != -1 {
				ext = serverFilename[idx:]
			}
		}
	}

	// Try URL path if no extension found
	if ext == "" {
		urlPath := resp.Request.URL.Path
		if idx := strings.LastIndex(urlPath, "."); idx != -1 {
			ext = urlPath[idx:]
		}
	}

	// If we got extension, add it to fallback
	if ext != "" && fallback != "" {
		return sanitizeFilename(fallback + ext)
	}

	// Last resort: use URL filename
	urlPath := resp.Request.URL.Path
	if idx := strings.LastIndex(urlPath, "/"); idx != -1 {
		filename := urlPath[idx+1:]
		if filename != "" && strings.Contains(filename, ".") {
			return sanitizeFilename(filename)
		}
	}

	// Default to fallback with .epub
	return sanitizeFilename(fallback + ".epub")
}

// sanitizeFilename removes unsafe characters from filename.
func sanitizeFilename(name string) string {
	// Remove path separators and other unsafe characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}
