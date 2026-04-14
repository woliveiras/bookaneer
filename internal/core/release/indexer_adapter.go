package release

import (
	"context"

	"github.com/woliveiras/bookaneer/internal/search"
)

// IndexerAdapter wraps a search.Indexer as a unified Source.
type IndexerAdapter struct {
	indexer search.Indexer
}

// NewIndexerAdapter creates a Source backed by a search.Indexer.
func NewIndexerAdapter(idx search.Indexer) *IndexerAdapter {
	return &IndexerAdapter{indexer: idx}
}

var _ Source = (*IndexerAdapter)(nil)

func (a *IndexerAdapter) Name() string     { return a.indexer.Name() }
func (a *IndexerAdapter) Type() SourceType { return SourceIndexer }

func (a *IndexerAdapter) Search(ctx context.Context, q Query) ([]Release, error) {
	sq := search.SearchQuery{
		Query:  q.Text,
		Author: q.Author,
		Title:  q.Title,
		ISBN:   q.ISBN,
		Limit:  q.Limit,
	}
	results, err := a.indexer.Search(ctx, sq)
	if err != nil {
		return nil, err
	}
	releases := make([]Release, 0, len(results))
	for _, r := range results {
		releases = append(releases, Release{
			ID:          r.GUID,
			Title:       r.Title,
			Size:        r.Size,
			DownloadURL: r.DownloadURL,
			InfoURL:     r.InfoURL,
			Provider:    r.IndexerName,
			SourceType:  SourceIndexer,
			Seeders:     r.Seeders,
			Leechers:    r.Leechers,
			Grabs:       r.Grabs,
			IndexerID:   r.IndexerID,
			IndexerName: r.IndexerName,
			Quality:     r.Quality,
		})
	}
	return releases, nil
}
