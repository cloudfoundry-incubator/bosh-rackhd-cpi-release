package models

// status of nodes that will be taged
const (
  Blocked        = "blocked"
  Unavailable    = "unavailable"
)

// Tags encapsulates a JSON of "tags" array for requests
type Tags struct {
  T []string `json:"tags"`
}

// TagNode is a node with and ID and an array of tags
type TagNode struct {
  ID   string   `json:"id"`
  Tags []string `json:"tags"`
  PersistentDisk PersistentDiskSettings `json:"persistent_disk"`
}

// PersistentDiskSettingsContainer is used to extract persistent_disk from tagnode
type PersistentDiskSettingsContainer struct {
  PersistentDisk PersistentDiskSettings `json:"persistent_disk"`
}

// PersistentDiskSettings is used to store value for persistent_disk
type PersistentDiskSettings struct {
  PregeneratedDiskCID string `json:"pregenerated_disk_cid"`
  DiskCID             string `json:"disk_cid"`
  Location            string `json:"location"`
  IsAttached          bool   `json:"attached"`
}

