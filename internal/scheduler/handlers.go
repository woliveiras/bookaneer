// Package scheduler provides handlers for command execution.
package scheduler

import (
	"context"
	"fmt"

	"github.com/woliveiras/bookaneer/internal/wanted"
)

// payloadFloat64 extracts a float64 value from a command payload map.
// JSON numbers are always decoded as float64.
func payloadFloat64(payload map[string]any, key string) (float64, bool) {
	v, ok := payload[key].(float64)
	return v, ok
}

// payloadString extracts a string value from a command payload map.
func payloadString(payload map[string]any, key string) (string, bool) {
	v, ok := payload[key].(string)
	return v, ok
}

// makeBookSearchHandler returns a CommandHandler that searches and grabs a specific book by ID.
// Used by both CommandBookSearch and CommandAutomaticSearch.
func makeBookSearchHandler(wantedSvc *wanted.Service) CommandHandler {
	return func(ctx context.Context, cmd *Command) error {
		bookID, ok := payloadFloat64(cmd.Payload, "bookId")
		if !ok {
			return fmt.Errorf("missing or invalid bookId in payload")
		}

		result, err := wantedSvc.SearchAndGrab(ctx, int64(bookID))
		if err != nil {
			return err
		}

		if result != nil {
			cmd.Result = map[string]any{
				"grabbed":  true,
				"title":    result.Title,
				"source":   result.Source,
				"provider": result.ProviderName,
				"format":   result.Format,
				"client":   result.ClientName,
			}
		} else {
			cmd.Result = map[string]any{
				"grabbed": false,
				"message": "No suitable release found",
			}
		}

		return nil
	}
}

// RegisterWantedHandlers registers handlers that use the wanted service.
func (s *Scheduler) RegisterWantedHandlers(wantedService *wanted.Service) {
	// BookSearch and AutomaticSearch share the same logic: search and grab by bookId.
	s.RegisterHandler(CommandBookSearch, makeBookSearchHandler(wantedService))
	s.RegisterHandler(CommandAutomaticSearch, makeBookSearchHandler(wantedService))

	// MissingBookSearch: Search all wanted (missing) books
	s.RegisterHandler(CommandMissingBookSearch, func(ctx context.Context, cmd *Command) error {
		results, err := wantedService.SearchAllWanted(ctx)
		if err != nil {
			return err
		}

		cmd.Result = map[string]any{
			"grabbed": len(results),
		}

		return nil
	})

	// DownloadGrab: Manually grab a specific release by URL
	s.RegisterHandler(CommandDownloadGrab, func(ctx context.Context, cmd *Command) error {
		bookID, ok := payloadFloat64(cmd.Payload, "bookId")
		if !ok {
			return fmt.Errorf("missing or invalid bookId in payload")
		}

		downloadURL, ok := payloadString(cmd.Payload, "downloadUrl")
		if !ok || downloadURL == "" {
			return fmt.Errorf("missing or invalid downloadUrl in payload")
		}

		releaseTitle, _ := payloadString(cmd.Payload, "releaseTitle")
		size, _ := payloadFloat64(cmd.Payload, "size")

		result, err := wantedService.GrabRelease(ctx, int64(bookID), downloadURL, releaseTitle, int64(size))
		if err != nil {
			return err
		}

		cmd.Result = map[string]any{
			"grabbed":    true,
			"downloadId": result.DownloadID,
			"client":     result.ClientName,
		}

		return nil
	})

	// RssSync: Periodically search for wanted books using RSS-enabled sources
	s.RegisterHandler(CommandRssSync, func(ctx context.Context, cmd *Command) error {
		results, err := wantedService.SearchAllWanted(ctx)
		if err != nil {
			return err
		}

		cmd.Result = map[string]any{
			"searched": true,
			"grabbed":  len(results),
			"message":  fmt.Sprintf("RSS sync completed, grabbed %d releases", len(results)),
		}

		return nil
	})

	// DownloadMonitor: Periodically check status of active downloads
	s.RegisterHandler(CommandDownloadMonitor, func(ctx context.Context, cmd *Command) error {
		result, err := wantedService.ProcessDownloads(ctx)
		if err != nil {
			return err
		}

		cmd.Result = map[string]any{
			"checked":   result.Checked,
			"completed": result.Completed,
			"failed":    result.Failed,
		}

		return nil
	})
}
