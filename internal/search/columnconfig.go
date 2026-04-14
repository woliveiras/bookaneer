package search

// ColumnRenderType controls how the frontend renders a column value.
type ColumnRenderType string

const (
	RenderText    ColumnRenderType = "text"
	RenderBadge   ColumnRenderType = "badge"
	RenderSize    ColumnRenderType = "size"
	RenderNumber  ColumnRenderType = "number"
	RenderPeers   ColumnRenderType = "peers"
	RenderIndexer ColumnRenderType = "indexer"
)

// ColumnAlign controls horizontal alignment.
type ColumnAlign string

const (
	AlignLeft   ColumnAlign = "left"
	AlignCenter ColumnAlign = "center"
	AlignRight  ColumnAlign = "right"
)

// ColumnColorHint tells the frontend how to color a badge column.
type ColumnColorHint struct {
	Type  string `json:"type"`  // "map" (use frontend color maps) or "static"
	Value string `json:"value"` // map name ("format", "language") or CSS class
}

// ColumnSchema defines a single column in the release list.
type ColumnSchema struct {
	Key        string           `json:"key"`
	Label      string           `json:"label"`
	RenderType ColumnRenderType `json:"renderType"`
	Align      ColumnAlign      `json:"align"`
	Width      string           `json:"width"`
	HideMobile bool             `json:"hideMobile"`
	ColorHint  *ColumnColorHint `json:"colorHint,omitempty"`
	Fallback   string           `json:"fallback"`
	Uppercase  bool             `json:"uppercase"`
	Sortable   bool             `json:"sortable"`
	SortKey    string           `json:"sortKey,omitempty"`
}

// ColumnConfig is the complete column configuration for a release source.
type ColumnConfig struct {
	Columns          []ColumnSchema `json:"columns"`
	GridTemplate     string         `json:"gridTemplate"`
	SupportedFilters []string       `json:"supportedFilters,omitempty"`
}

// LibraryColumnConfig returns the default column configuration for digital library results.
func LibraryColumnConfig() ColumnConfig {
	return ColumnConfig{
		Columns: []ColumnSchema{
			{
				Key:        "format",
				Label:      "Format",
				RenderType: RenderBadge,
				Align:      AlignCenter,
				Width:      "70px",
				ColorHint:  &ColumnColorHint{Type: "map", Value: "format"},
				Fallback:   "-",
				Uppercase:  true,
			},
			{
				Key:        "language",
				Label:      "Language",
				RenderType: RenderBadge,
				Align:      AlignCenter,
				Width:      "60px",
				HideMobile: true,
				ColorHint:  &ColumnColorHint{Type: "map", Value: "language"},
				Fallback:   "-",
				Uppercase:  true,
			},
			{
				Key:        "size",
				Label:      "Size",
				RenderType: RenderSize,
				Align:      AlignCenter,
				Width:      "80px",
				Fallback:   "-",
			},
			{
				Key:        "provider",
				Label:      "Source",
				RenderType: RenderText,
				Align:      AlignLeft,
				Width:      "120px",
				HideMobile: true,
				Fallback:   "-",
			},
		},
		GridTemplate:     "minmax(0,2fr) 70px 60px 80px 120px",
		SupportedFilters: []string{"format", "language"},
	}
}

// IndexerColumnConfig returns the default column configuration for indexer results.
func IndexerColumnConfig() ColumnConfig {
	return ColumnConfig{
		Columns: []ColumnSchema{
			{
				Key:        "indexerName",
				Label:      "Indexer",
				RenderType: RenderIndexer,
				Align:      AlignLeft,
				Width:      "minmax(100px, 1fr)",
			},
			{
				Key:        "seeders",
				Label:      "S/L",
				RenderType: RenderPeers,
				Align:      AlignCenter,
				Width:      "70px",
				Sortable:   true,
				SortKey:    "seeders",
			},
			{
				Key:        "size",
				Label:      "Size",
				RenderType: RenderSize,
				Align:      AlignCenter,
				Width:      "80px",
				Sortable:   true,
				SortKey:    "size",
			},
		},
		GridTemplate: "minmax(0,2fr) minmax(100px,1fr) 70px 80px",
	}
}
