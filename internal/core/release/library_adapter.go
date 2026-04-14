package release

import (
	"context"

	"github.com/woliveiras/bookaneer/internal/library"
)

// LibraryAdapter wraps a library.Provider as a unified Source.
type LibraryAdapter struct {
	provider library.Provider
}

// NewLibraryAdapter creates a Source backed by a library.Provider.
func NewLibraryAdapter(p library.Provider) *LibraryAdapter {
	return &LibraryAdapter{provider: p}
}

var _ Source = (*LibraryAdapter)(nil)

func (a *LibraryAdapter) Name() string     { return a.provider.Name() }
func (a *LibraryAdapter) Type() SourceType { return SourceLibrary }

func (a *LibraryAdapter) Search(ctx context.Context, q Query) ([]Release, error) {
	text := q.Text
	if text == "" {
		text = q.Title
	}
	results, err := a.provider.Search(ctx, text)
	if err != nil {
		return nil, err
	}
	releases := make([]Release, 0, len(results))
	for _, r := range results {
		releases = append(releases, Release{
			ID:          r.ID,
			Title:       r.Title,
			Authors:     r.Authors,
			Format:      r.Format,
			Size:        r.Size,
			DownloadURL: r.DownloadURL,
			InfoURL:     r.InfoURL,
			Provider:    r.Provider,
			SourceType:  SourceLibrary,
			Language:    r.Language,
			Year:        r.Year,
			ISBN:        r.ISBN,
			CoverURL:    r.CoverURL,
			Score:       r.Score,
		})
	}
	return releases, nil
}
