package models

// Metadata in tags as "metadata:JSON STRUCT"
type Metadata struct {
	ID        string `json:"id"`
	Workflows string `json:"workflows"`
	OBMS      []OBM  `json:"obms"`
	CID       string `json:"cid"`
	Status    string `json:"status"`
}

// PersistentDiskSettings in tags as "persistentdisksettings:JSON STRUCT"
type PersistentDiskSettings struct {
	PregeneratedDiskCID string `json:"pregenerated_disk_cid"`
	DiskCID             string `json:"disk_cid"`
	Location            string `json:"location"`
	IsAttached          bool   `json:"attached"`
}

const (
	NetworkActive    = "up"
	NetworkInactive  = "down"
	EthernetNetwork  = "Ethernet"
	MacAddressFamily = "lladdr"
)

const (
	Available   = "available"
	Reserved    = "reserved"
	Blocked     = "blocked"
	DiskReason  = "Node has missing disks"
	Maintenance = "maintenance"
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
