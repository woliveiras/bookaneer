// Package scheduler provides handlers for command execution.
package scheduler

import (
	"context"
	"fmt"

	"github.com/woliveiras/bookaneer/internal/wanted"
)

// RegisterWantedHandlers registers handlers that use the wanted service.
func (s *Scheduler) RegisterWantedHandlers(wantedService *wanted.Service) {
	// BookSearch: Search and grab a specific book by ID
	s.RegisterHandler(CommandBookSearch, func(ctx context.Context, cmd *Command) error {
		bookID, ok := cmd.Payload["bookId"].(float64) // JSON numbers are float64
		if !ok {
			return fmt.Errorf("missing or invalid bookId in payload")
		}

		result, err := wantedService.SearchAndGrab(ctx, int64(bookID))
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
	})

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
}
