package trash

import "time"

// ItemType represents the type of a trashed item.
type ItemType string

const (
	ItemTypeFile      ItemType = "file"
	ItemTypeDirectory ItemType = "directory"
)

// Item represents a single trashed item.
type Item struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	OriginalPath string   `json:"original_path"`
	TrashPath    string   `json:"trash_path"`
	Type         ItemType `json:"item_type"`
	SizeBytes    int64    `json:"size_bytes"`
	TrashedAt    string   `json:"trashed_at"`
	Metadata     string   `json:"metadata,omitempty"`
}

// ListFilter controls filtering for the List method.
type ListFilter struct {
	Type      ItemType      // Filter by item type (empty = all)
	OlderThan time.Duration // Filter items older than this duration (zero = all)
}
