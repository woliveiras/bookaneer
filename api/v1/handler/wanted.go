package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

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
	// Wanted books
	g.GET("/wanted/missing", h.GetMissingBooks)
	g.POST("/wanted/missing/search", h.SearchAllMissing)
	g.POST("/wanted/cutoff", h.GetCutoffUnmet)
	g.POST("/wanted/cutoff/search", h.SearchCutoffUnmet)

	// Download queue
	g.GET("/queue", h.GetQueue)
	g.DELETE("/queue/:id", h.RemoveFromQueue)

	// Active commands (for Activity page)
	g.GET("/commands/active", h.GetActiveCommands)

	// Manual search and grab
	g.POST("/book/:id/search", h.SearchBook)
	g.POST("/release", h.ManualGrab)
}

// GetMissingBooks returns all monitored books without files.
func (h *WantedHandler) GetMissingBooks(c echo.Context) error {
	books, err := h.wantedService.GetWantedBooks(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get missing books")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"page":          1,
		"pageSize":      len(books),
		"totalRecords":  len(books),
		"sortKey":       "addedAt",
		"sortDirection": "descending",
		"records":       books,
	})
}

// SearchAllMissing triggers a search for all missing books.
func (h *WantedHandler) SearchAllMissing(c echo.Context) error {
	ctx := c.Request().Context()

	// Queue a MissingBookSearch command
	commandID, err := h.schedulerService.QueueCommand(ctx, scheduler.CommandMissingBookSearch, scheduler.TriggerManual, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to queue search command")
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"commandId": commandID,
		"message":   "Search for all missing books has been queued",
	})
}

// GetCutoffUnmet returns books that don't meet quality cutoff.
// For now, returns empty as quality cutoff is not yet implemented.
func (h *WantedHandler) GetCutoffUnmet(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"page":          1,
		"pageSize":      0,
		"totalRecords":  0,
		"sortKey":       "addedAt",
		"sortDirection": "descending",
		"records":       []any{},
	})
}

// SearchCutoffUnmet triggers a search for cutoff unmet books.
func (h *WantedHandler) SearchCutoffUnmet(c echo.Context) error {
	return c.JSON(http.StatusAccepted, map[string]any{
		"message": "Cutoff search is not yet implemented",
	})
}

// GetQueue returns the current download queue.
func (h *WantedHandler) GetQueue(c echo.Context) error {
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
func (h *WantedHandler) RemoveFromQueue(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid queue item id")
	}

	if err := h.wantedService.RemoveFromQueue(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to remove from queue")
	}

	return c.NoContent(http.StatusOK)
}

// GetActiveCommands returns commands that are queued or running.
func (h *WantedHandler) GetActiveCommands(c echo.Context) error {
	commands, err := h.schedulerService.GetActiveCommands(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get active commands")
	}

	if commands == nil {
		commands = []scheduler.Command{}
	}

	return c.JSON(http.StatusOK, commands)
}

// SearchBookRequest represents a manual book search request.
type SearchBookRequest struct {
	BookID int64 `json:"bookId"`
}

// SearchBook triggers a search for a specific book.
func (h *WantedHandler) SearchBook(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book id")
	}

	ctx := c.Request().Context()

	// Queue a BookSearch command
	commandID, err := h.schedulerService.QueueCommand(ctx, scheduler.CommandBookSearch, scheduler.TriggerManual, map[string]any{
		"bookId": id,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to queue search command")
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"commandId": commandID,
		"message":   "Search has been queued",
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

// ManualGrab manually grabs a release.
func (h *WantedHandler) ManualGrab(c echo.Context) error {
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

	// Queue a DownloadGrab command
	ctx := c.Request().Context()
	commandID, err := h.schedulerService.QueueCommand(ctx, scheduler.CommandDownloadGrab, scheduler.TriggerManual, map[string]any{
		"bookId":       req.BookID,
		"indexerId":    req.IndexerID,
		"downloadUrl":  req.DownloadURL,
		"releaseTitle": req.ReleaseTitle,
		"size":         req.Size,
		"quality":      req.Quality,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to queue grab command")
	}

	return c.JSON(http.StatusAccepted, map[string]any{
		"commandId": commandID,
		"message":   "Grab has been queued",
	})
}
