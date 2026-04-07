package qualityprofile

import "encoding/json"

// QualityProfile represents a quality profile for ebook downloads.
type QualityProfile struct {
	ID     int64         `json:"id"`
	Name   string        `json:"name"`
	Cutoff string        `json:"cutoff"`
	Items  []QualityItem `json:"items"`
}

// QualityItem represents a quality setting within a profile.
type QualityItem struct {
	Quality string `json:"quality"`
	Allowed bool   `json:"allowed"`
}

// CreateQualityProfileInput holds the data needed to create a new quality profile.
type CreateQualityProfileInput struct {
	Name   string        `json:"name"`
	Cutoff string        `json:"cutoff"`
	Items  []QualityItem `json:"items"`
}

// UpdateQualityProfileInput holds the data for updating an existing quality profile.
type UpdateQualityProfileInput struct {
	Name   *string        `json:"name,omitempty"`
	Cutoff *string        `json:"cutoff,omitempty"`
	Items  *[]QualityItem `json:"items,omitempty"`
}

// MarshalItems converts items to JSON string for storage.
func MarshalItems(items []QualityItem) string {
	data, _ := json.Marshal(items)
	return string(data)
}

// UnmarshalItems converts JSON string to items.
func UnmarshalItems(data string) []QualityItem {
	var items []QualityItem
	_ = json.Unmarshal([]byte(data), &items)
	return items
}

// DefaultProfile returns the default quality profile.
func DefaultProfile() *QualityProfile {
	return &QualityProfile{
		Name:   "Default",
		Cutoff: "epub",
		Items: []QualityItem{
			{Quality: "epub", Allowed: true},
			{Quality: "mobi", Allowed: true},
			{Quality: "azw3", Allowed: true},
			{Quality: "pdf", Allowed: false},
			{Quality: "cbz", Allowed: false},
		},
	}
}

// AllQualities returns all supported quality formats.
func AllQualities() []string {
	return []string{"epub", "mobi", "azw3", "pdf", "cbz"}
}
