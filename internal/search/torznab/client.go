package torznab

import (
	"github.com/woliveiras/bookaneer/internal/search"
	"github.com/woliveiras/bookaneer/internal/search/newznab"
)

func init() {
	search.RegisterFactory("torznab", func(cfg search.IndexerConfig) (search.Indexer, error) {
		return New(cfg), nil
	})
}

// Client implements the Torznab API (extends Newznab).
type Client struct {
	*newznab.Client
}

// New creates a new Torznab client.
func New(cfg search.IndexerConfig) *Client {
	return &Client{
		Client: newznab.New(cfg),
	}
}

func (c *Client) Type() string { return "torznab" }
