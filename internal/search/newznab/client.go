package newznab

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/search"
)

func init() {
	search.RegisterFactory("newznab", func(cfg search.IndexerConfig) (search.Indexer, error) {
		return New(cfg), nil
	})
}

// Client implements the Newznab API.
type Client struct {
	config     search.IndexerConfig
	httpClient *http.Client
}

// New creates a new Newznab client.
func New(cfg search.IndexerConfig) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Name() string { return c.config.Name }
func (c *Client) Type() string { return "newznab" }

// Search searches the indexer for releases.
func (c *Client) Search(ctx context.Context, query search.SearchQuery) ([]search.Result, error) {
	params := url.Values{
		"t":      {"search"},
		"apikey": {c.config.APIKey},
		"o":      {"json"},
	}
	if query.Query != "" {
		params.Set("q", query.Query)
	}
	if query.Author != "" {
		params.Set("author", query.Author)
	}
	if query.Title != "" {
		params.Set("title", query.Title)
	}
	if query.Limit > 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}
	if query.Offset > 0 {
		params.Set("offset", strconv.Itoa(query.Offset))
	}
	if len(query.Category) > 0 {
		params.Set("cat", strings.Join(query.Category, ","))
	}
	u := fmt.Sprintf("%s%s?%s", strings.TrimRight(c.config.BaseURL, "/"), c.config.APIPath, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, search.ErrInvalidAPIKey
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, search.ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return c.parseRSS(body)
}

type rss struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Size        int64  `xml:"size"`
	Category    string `xml:"category"`
	Attrs       []attr `xml:"attr"`
}

type attr struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func (c *Client) parseRSS(data []byte) ([]search.Result, error) {
	var feed rss
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("unmarshal rss: %w", err)
	}
	results := make([]search.Result, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		r := search.Result{
			GUID:        item.GUID,
			Title:       item.Title,
			Description: item.Description,
			Size:        item.Size,
			DownloadURL: item.Link,
			Category:    item.Category,
			IndexerID:   c.config.ID,
			IndexerName: c.config.Name,
		}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			r.PubDate = t
		}
		for _, a := range item.Attrs {
			switch a.Name {
			case "grabs":
				r.Grabs, _ = strconv.Atoi(a.Value)
			case "seeders":
				r.Seeders, _ = strconv.Atoi(a.Value)
			case "leechers":
				r.Leechers, _ = strconv.Atoi(a.Value)
			case "comments":
				r.Comments, _ = strconv.Atoi(a.Value)
			case "category":
				r.CategoryID = a.Value
			}
		}
		results = append(results, r)
	}
	return results, nil
}

// Caps retrieves the capabilities of the indexer.
func (c *Client) Caps(ctx context.Context) (*search.Capabilities, error) {
	u := fmt.Sprintf("%s%s?t=caps&apikey=%s", strings.TrimRight(c.config.BaseURL, "/"), c.config.APIPath, url.QueryEscape(c.config.APIKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, search.ErrInvalidAPIKey
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return c.parseCaps(body)
}

type capsXML struct {
	Searching struct {
		Search     capsSearchAttr `xml:"search"`
		BookSearch capsSearchAttr `xml:"book-search"`
	} `xml:"searching"`
	Categories struct {
		Category []capsCat `xml:"category"`
	} `xml:"categories"`
}

type capsSearchAttr struct {
	Available string `xml:"available,attr"`
}

type capsCat struct {
	ID     string    `xml:"id,attr"`
	Name   string    `xml:"name,attr"`
	SubCat []capsCat `xml:"subcat"`
}

func (c *Client) parseCaps(data []byte) (*search.Capabilities, error) {
	var caps capsXML
	if err := xml.Unmarshal(data, &caps); err != nil {
		return nil, fmt.Errorf("unmarshal caps: %w", err)
	}
	result := &search.Capabilities{}
	result.Searching.Search = caps.Searching.Search.Available == "yes"
	result.Searching.BookSearch = caps.Searching.BookSearch.Available == "yes"
	result.Categories = make([]search.Category, 0, len(caps.Categories.Category))
	for _, cat := range caps.Categories.Category {
		c := search.Category{ID: cat.ID, Name: cat.Name}
		for _, sub := range cat.SubCat {
			c.SubCategory = append(c.SubCategory, search.Category{ID: sub.ID, Name: sub.Name})
		}
		result.Categories = append(result.Categories, c)
	}
	return result, nil
}

// Test tests the connection to the indexer.
func (c *Client) Test(ctx context.Context) error {
	_, err := c.Caps(ctx)
	return err
}
