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

// RegisterWantedHandlers registers handlers that use the wanted service.
func (s *Scheduler) RegisterWantedHandlers(wantedService *wanted.Service) {
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
