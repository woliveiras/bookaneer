package wanted

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/woliveiras/bookaneer/internal/download"
)

// GetDownloadQueue returns the current download queue.
func (s *Service) GetDownloadQueue(ctx context.Context) ([]DownloadQueueItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT dq.id, dq.book_id, dq.download_client_id, dq.indexer_id, dq.external_id,
		       dq.title, dq.size, dq.format, dq.status, dq.progress, dq.download_url, dq.added_at,
		       b.title as book_title,
		       dc.name as client_name,
		       dq.error_message
		FROM download_queue dq
		LEFT JOIN books b ON b.id = dq.book_id
		LEFT JOIN download_clients dc ON dc.id = dq.download_client_id
		ORDER BY dq.added_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []DownloadQueueItem
	for rows.Next() {
		var item DownloadQueueItem
		var clientID sql.NullInt64
		var indexerID sql.NullInt64
		var bookTitle sql.NullString
		var clientName sql.NullString
		var errorMessage sql.NullString
		if err := rows.Scan(
			&item.ID, &item.BookID, &clientID, &indexerID, &item.ExternalID,
			&item.Title, &item.Size, &item.Format, &item.Status, &item.Progress, &item.DownloadURL, &item.AddedAt,
			&bookTitle, &clientName, &errorMessage,
		); err != nil {
			return nil, err
		}
		if clientID.Valid {
			item.DownloadClientID = &clientID.Int64
		}
		if indexerID.Valid {
			item.IndexerID = &indexerID.Int64
		}
		if bookTitle.Valid {
			item.BookTitle = bookTitle.String
		} else {
			item.BookTitle = item.Title // Fallback to release title
		}
		if clientName.Valid {
			item.ClientName = clientName.String
		} else {
			item.ClientName = "Embedded Downloader"
		}
		if errorMessage.Valid {
			item.ErrorMessage = errorMessage.String
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// For embedded client items, get real-time status from the direct client
	client, _, err := s.downloadService.GetDirectClient(ctx)
	if err == nil && client != nil {
		for i := range items {
			// Check items with no client ID (embedded) that have active statuses
			if items[i].DownloadClientID == nil && items[i].ExternalID != "" {
				status, err := client.GetStatus(ctx, items[i].ExternalID)
				if err == nil {
					// Update with real-time status from embedded client
					items[i].Status = string(status.Status)
					items[i].Progress = status.Progress
					items[i].ErrorMessage = status.ErrorMessage
					// Also update DB to persist the status and error
					if status.ErrorMessage != "" {
						_ = s.UpdateQueueItemStatusWithError(ctx, items[i].ID, items[i].Status, items[i].Progress, status.ErrorMessage)
					} else {
						_ = s.UpdateQueueItemStatus(ctx, items[i].ID, items[i].Status, items[i].Progress)
					}
				}
			}
		}
	}

	return items, nil
}

// UpdateQueueItemStatus updates the status of a queue item.
func (s *Service) UpdateQueueItemStatus(ctx context.Context, id int64, status string, progress float64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE download_queue SET status = ?, progress = ? WHERE id = ?`, status, progress, id)
	return err
}

// UpdateQueueItemStatusWithError updates status and persists the error message.
func (s *Service) UpdateQueueItemStatusWithError(ctx context.Context, id int64, status string, progress float64, errorMessage string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE download_queue SET status = ?, progress = ?, error_message = ? WHERE id = ?`, status, progress, errorMessage, id)
	return err
}

// UpdateQueueItemStatusWithPath updates the status and save_path of a queue item.
func (s *Service) UpdateQueueItemStatusWithPath(ctx context.Context, id int64, status string, progress float64, savePath string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE download_queue SET status = ?, progress = ?, save_path = ? WHERE id = ?`, status, progress, savePath, id)
	return err
}

// RemoveFromQueue removes an item from the download queue.
func (s *Service) RemoveFromQueue(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM download_queue WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete query failed: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("queue item %d not found", id)
	}
	return nil
}

// RetryDownload re-submits a failed or cancelled download to the download client.
func (s *Service) RetryDownload(ctx context.Context, id int64) error {
	client, _, err := s.downloadService.GetDirectClient(ctx)
	if err != nil || client == nil {
		return fmt.Errorf("get download client: %w", err)
	}
	return s.restartDownload(ctx, id, client)
}

// recordDownload adds an entry to the download_queue table.
// clientID can be nil for embedded client (no database entry).
func (s *Service) recordDownload(ctx context.Context, bookID int64, clientID *int64, indexerID *int64, title string, size int64, format, downloadURL, externalID, savePath string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO download_queue (book_id, download_client_id, indexer_id, external_id, title, size, format, status, download_url, save_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'queued', ?, ?)
	`, bookID, clientID, indexerID, externalID, title, size, format, downloadURL, savePath)
	return err
}

// restartDownload restarts a download that was lost (e.g., after server restart).
func (s *Service) restartDownload(ctx context.Context, queueID int64, client download.Client) error {
	// Get download info from queue
	var title, downloadURL string
	err := s.db.QueryRowContext(ctx, `
		SELECT title, download_url FROM download_queue WHERE id = ?
	`, queueID).Scan(&title, &downloadURL)
	if err != nil {
		return fmt.Errorf("get queue item: %w", err)
	}

	if downloadURL == "" {
		return fmt.Errorf("no download URL for queue item %d", queueID)
	}

	// Add to client again
	newID, err := client.Add(ctx, download.AddItem{
		Name:        title,
		DownloadURL: downloadURL,
		Category:    "books",
	})
	if err != nil {
		return fmt.Errorf("add to client: %w", err)
	}

	// Update external_id in queue
	_, err = s.db.ExecContext(ctx, `UPDATE download_queue SET external_id = ?, status = 'queued' WHERE id = ?`, newID, queueID)
	if err != nil {
		return fmt.Errorf("update queue: %w", err)
	}

	slog.Info("Download restarted", "queueId", queueID, "newExternalId", newID)
	return nil
}
