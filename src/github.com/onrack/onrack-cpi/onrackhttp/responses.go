package onrackhttp

const (
	NetworkActive    = "up"
	NetworkInactive  = "down"
	EthernetNetwork  = "Ethernet"
	MacAddressFamily = "lladdr"
)

type FileMetadataResponse []struct {
	Basename string `json:"basename"`
	Filename string `json:"filename"`
	UUID     string `json:"uuid"`
	Md5      string `json:"md5"`
	Sha256   string `json:"sha256"`
	Version  int    `json:"version"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type NodeCatalog struct {
	Data CatalogData `json:"data"`
}

type CatalogData struct {
	NetworkData NetworkCatalog `json:"network"`
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

type Node struct {
	Workflows []interface{} `json:"workflows"`
	Reserved  string        `json:"reserved"`
	ID        string        `json:"id"`
	CID       string        `json:"cid"`
}
