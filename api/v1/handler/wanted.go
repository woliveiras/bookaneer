package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/woliveiras/bookaneer/internal/scheduler"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

// WantedHandler handles wanted books and download queue HTTP requests.
type WantedHandler struct {
	wantedService    *wanted.Service
	schedulerService *scheduler.Scheduler
}

// NewWantedHandler creates a new wanted handler.
func NewWantedHandler(wantedSvc *wanted.Service, schedulerSvc *scheduler.Scheduler) *WantedHandler {
	return &WantedHandler{
		wantedService:    wantedSvc,
		schedulerService: schedulerSvc,
	}
}

// Register registers the wanted routes.
func (h *WantedHandler) Register(g *echo.Group) {
	// Download queue
	g.GET("/queue", h.GetQueue)
	g.DELETE("/queue/:id", h.RemoveFromQueue)
	g.POST("/queue/:id/retry", h.RetryDownload)

	// History
	g.GET("/history", h.GetHistory)

	// Blocklist
	g.GET("/blocklist", h.GetBlocklist)
	g.POST("/blocklist", h.AddToBlocklist)
	g.DELETE("/blocklist/:id", h.RemoveFromBlocklist)

	// Commands (for Activity and System pages)
	g.GET("/commands/active", h.GetActiveCommands)
	g.GET("/commands/recent", h.GetRecentCommands)

	// Manual search and grab
	g.POST("/book/:id/search", h.SearchBook)
	g.POST("/book/:id/wrong-content", h.ReportWrongContent)
	g.POST("/release", h.ManualGrab)
	g.POST("/indexer-release", h.GrabIndexerRelease)
}

// GetQueue returns the current download queue.
func (h *WantedHandler) GetQueue(c *echo.Context) error {
	queue, err := h.wantedService.GetDownloadQueue(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get download queue")
	}

	if queue == nil {
		queue = []wanted.DownloadQueueItem{}
	}

	return c.JSON(http.StatusOK, queue)
}

// RemoveFromQueue removes an item from the download queue.
func (h *WantedHandler) RemoveFromQueue(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid queue item id")
	}

	if err := h.wantedService.RemoveFromQueue(c.Request().Context(), id); err != nil {
		slog.Error("failed to remove queue item", "id", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove from queue")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetActiveCommands returns commands that are queued or running.
func (h *WantedHandler) GetActiveCommands(c *echo.Context) error {
	commands, err := h.schedulerService.GetActiveCommands(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get active commands")
	}

	if commands == nil {
		commands = []scheduler.Command{}
	}

	return c.JSON(http.StatusOK, commands)
}

// GetRecentCommands returns the most recent commands for the logs view.
func (h *WantedHandler) GetRecentCommands(c *echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	commands, err := h.schedulerService.ListCommands(c.Request().Context(), limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get recent commands")
	}

	if commands == nil {
		commands = []scheduler.Command{}
	}

	return c.JSON(http.StatusOK, commands)
}

// SearchBook searches for releases for a specific book and returns the results.
// Results are sorted by file size (largest first). If no results are found, the
// empty results list is returned with noResults=true so the client can offer to
// add the book to the wanted list.
func (h *WantedHandler) SearchBook(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	ctx := c.Request().Context()

	results, err := h.wantedService.Search(ctx, id)
	if err != nil {
		if err.Error() == "find book: not found" {
			return echo.NewHTTPError(http.StatusNotFound, "book not found")
		}
		slog.Error("search failed", "bookId", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "search failed")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"results":   results,
		"noResults": len(results) == 0,
	})
}

// ManualGrabRequest represents a manual grab request.
type ManualGrabRequest struct {
	BookID       int64  `json:"bookId"`
	IndexerID    int64  `json:"indexerId,omitempty"`
	DownloadURL  string `json:"downloadUrl"`
	ReleaseTitle string `json:"releaseTitle"`
	Size         int64  `json:"size"`
	Quality      string `json:"quality"`
}

// IndexerGrabRequest represents an indexer grab request.
type IndexerGrabRequest struct {
	BookID       int64  `json:"bookId"`
	GUID         string `json:"guid"`
	DownloadURL  string `json:"downloadUrl"`
	ReleaseTitle string `json:"releaseTitle"`
	Size         int64  `json:"size"`
	Seeders      int    `json:"seeders"`
	IndexerID    int64  `json:"indexerId"`
	IndexerName  string `json:"indexerName"`
}

// ManualGrab immediately grabs a release and starts the download.
func (h *WantedHandler) ManualGrab(c *echo.Context) error {
	var req ManualGrabRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.BookID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "bookId is required")
	}
	if req.DownloadURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "downloadUrl is required")
	}

	ctx := c.Request().Context()
	result, err := h.wantedService.GrabRelease(ctx, req.BookID, req.DownloadURL, req.ReleaseTitle, req.Size)
	if err != nil {
		slog.Error("grab failed", "bookId", req.BookID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to start download")
	}

	return c.JSON(http.StatusOK, result)
}

// GrabIndexerRelease grabs an indexer result and routes it to the appropriate torrent or usenet client.
func (h *WantedHandler) GrabIndexerRelease(c *echo.Context) error {
	var req IndexerGrabRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.BookID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "bookId is required")
	}
	if req.DownloadURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "downloadUrl is required")
	}

	ctx := c.Request().Context()
	result, err := h.wantedService.GrabIndexerRelease(ctx, req.BookID, wanted.GrabIndexerRequest{
		GUID:         req.GUID,
		ReleaseTitle: req.ReleaseTitle,
		DownloadURL:  req.DownloadURL,
		Size:         req.Size,
		Seeders:      req.Seeders,
		IndexerID:    req.IndexerID,
		IndexerName:  req.IndexerName,
	})
	if err != nil {
		slog.Error("indexer grab failed", "bookId", req.BookID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to start download")
	}

	return c.JSON(http.StatusOK, result)
}

// RetryDownload retries a failed or cancelled download queue item.
func (h *WantedHandler) RetryDownload(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid queue item id")
	}

	ctx := c.Request().Context()
	if err := h.wantedService.RetryDownload(ctx, id); err != nil {
		slog.Error("retry failed", "queueId", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retry download")
	}

	return c.NoContent(http.StatusNoContent)
}

// GetHistory returns history events.
func (h *WantedHandler) GetHistory(c *echo.Context) error {
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	eventType := c.QueryParam("eventType")

	items, err := h.wantedService.GetHistory(c.Request().Context(), limit, eventType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get history")
	}

	if items == nil {
		items = []wanted.HistoryItem{}
	}

	return c.JSON(http.StatusOK, items)
}

// GetBlocklist returns blocklisted releases.
func (h *WantedHandler) GetBlocklist(c *echo.Context) error {
	items, err := h.wantedService.GetBlocklist(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get blocklist")
	}

	if items == nil {
		items = []wanted.BlocklistItem{}
	}

	return c.JSON(http.StatusOK, items)
}

// AddToBlocklistRequest represents a request to add to blocklist.
type AddToBlocklistRequest struct {
	BookID      int64  `json:"bookId"`
	SourceTitle string `json:"sourceTitle"`
	Quality     string `json:"quality"`
	Reason      string `json:"reason"`
}

// AddToBlocklist adds a release to the blocklist.
func (h *WantedHandler) AddToBlocklist(c *echo.Context) error {
	var req AddToBlocklistRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.BookID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "bookId is required")
	}
	if req.SourceTitle == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sourceTitle is required")
	}

	if err := h.wantedService.AddToBlocklist(c.Request().Context(), req.BookID, req.SourceTitle, req.Quality, req.Reason); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add to blocklist")
	}

	return c.NoContent(http.StatusCreated)
}

// RemoveFromBlocklist removes an item from the blocklist.
func (h *WantedHandler) RemoveFromBlocklist(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid blocklist item id")
	}

	if err := h.wantedService.RemoveFromBlocklist(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove from blocklist")
	}

	return c.NoContent(http.StatusOK)
}

// ReportWrongContent marks a book file as having wrong content.
// It removes the file, blocklists the source, and tries the next available source.
func (h *WantedHandler) ReportWrongContent(c *echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	if err := h.wantedService.ReportWrongContent(c.Request().Context(), id); err != nil {
		slog.Error("failed to report wrong content", "bookId", id, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to report wrong content")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "File removed, source blocklisted. Searching for alternative sources.",
	})
}
