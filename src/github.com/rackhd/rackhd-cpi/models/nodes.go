package models

const (
  NetworkActive    = "up"
  NetworkInactive  = "down"
  EthernetNetwork  = "Ethernet"
  MacAddressFamily = "lladdr"
)

const (
  PersistentDiskLocation = "sdb"
)

type NodeCatalog struct {
  Data CatalogData `json:"data"`
}

type Device struct {
  Size string `json:"size"`
}

type CatalogData struct {
  NetworkData  NetworkCatalog    `json:"network"`
  BlockDevices map[string]Device `json:"block_device"`
}

type NetworkCatalog struct {
  Networks map[string]Network `json:"interfaces"`
}

type Network struct {
  Encapsulation string                    `json:"encapsulation"`
  Number        string                    `json:"number"`
  Addresses     map[string]NetworkAddress `json:"addresses"`
  State         string                    `json:"state"`
}

type NetworkAddress struct {
  Family string `json:"family"`
}

type OBM struct {
  ServiceName string `json:"service"`
  Ref         string `json:"ref"`
}

type PersistentDiskSettingsContainer struct {
  PersistentDisk PersistentDiskSettings `json:"persistent_disk"`
}

type PersistentDiskSettings struct {
  PregeneratedDiskCID string `json:"pregenerated_disk_cid"`
  DiskCID             string `json:"disk_cid"`
  Location            string `json:"location"`
  IsAttached          bool   `json:"attached"`
}

type Node struct {
  ID             string                 `json:"id"`
  Workflows      string                 `json:"workflows"`
  OBMS           []OBM                  `json:"obms"`
  PersistentDisk PersistentDiskSettings `json:"persistent_disk"`
}
