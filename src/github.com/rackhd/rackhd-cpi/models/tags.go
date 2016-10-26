package models

const (
  Blocked  = "blocked"
  Reserved = "reserved"
)

type Tags struct {
  T []string `json:"tags"`
}

type TagNode struct {
  ID string `json:"id"`
  Tags
}
