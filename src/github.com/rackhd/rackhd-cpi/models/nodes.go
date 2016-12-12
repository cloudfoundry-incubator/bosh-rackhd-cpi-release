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

type BMCCatalog struct {
  Data BMCCatalogData `json:"data"`
}

type BMCCatalogData struct {
  MACAddress string `json:"MAC Address"`
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

type OBMServiceRequest struct {
  Config      OBMConfig `json:"config"`
  NodeID      string    `json:"nodeId"`
  ServiceName string    `json:"service"`
}

type OBMConfig struct {
  Host     string `json:"host"`
  Password string `json:"password"`
  User     string `json:"user"`
}

type Node struct {
  ID             string                 `json:"id"`
  Name           string                 `json:"name"`
  Workflows      string                 `json:"workflows"`
  OBMS           []OBM                  `json:"obms"`
}
