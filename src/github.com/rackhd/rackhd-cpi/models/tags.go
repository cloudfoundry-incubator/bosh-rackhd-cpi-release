package models

const (
	Blocked        = "blocked"
	Reserved       = "reserved"
	Available      = "available"
	PersistentDisk = "diskprovisioned"
)

// Tags encapsulates a JSON of "tags" array for requests
type Tags struct {
	T []string `json:"tags"`
}

// TagNode is a node with and ID and an array of tags
type TagNode struct {
	ID   string   `json:"id"`
	Tags []string `json:"tags"`
}
