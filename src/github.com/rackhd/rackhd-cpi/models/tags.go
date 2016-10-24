package models

const (
  Available   = "available"
  Unavailable = "unavailable"
)

type Tags struct {
  T []string `json:"tags"`
}

type TagNode struct {
  ID string `json:"id"`
  Tags
}
