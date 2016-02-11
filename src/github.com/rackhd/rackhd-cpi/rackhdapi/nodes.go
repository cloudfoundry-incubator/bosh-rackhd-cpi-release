package rackhdapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rackhd/rackhd-cpi/config"
)

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

type OBMSetting struct {
	Config      interface{} `json:"config"`
	ServiceName string      `json:"service"`
}

type PersistentDiskSettingsContainer struct {
	PersistentDisk PersistentDiskSettings `json:"persistent_disk"`
}

type PersistentDiskSettings struct {
	DiskCID    string `json:"disk_cid"`
	Location   string `json:"location"`
	IsAttached bool   `json:"attached"`
}

type Node struct {
	Workflows      []interface{}          `json:"workflows"`
	Status         string                 `json:"status"`
	ID             string                 `json:"id"`
	CID            string                 `json:"cid"`
	OBMSettings    []OBMSetting           `json:"obmSettings"`
	PersistentDisk PersistentDiskSettings `json:"persistent_disk"`
}

func GetNodes(c config.Cpi) ([]Node, error) {
	nodesURL := fmt.Sprintf("http://%s/api/common/nodes", c.ApiServer)
	resp, err := http.Get(nodesURL)
	if err != nil {
		return []Node{}, fmt.Errorf("error fetching nodes %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []Node{}, fmt.Errorf("Failed getting nodes with status: %s, err: %s", resp.Status, err)
	}

	nodeBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Node{}, fmt.Errorf("error reading node response body %s", err)
	}

	var nodes []Node
	err = json.Unmarshal(nodeBytes, &nodes)
	if err != nil {
		return []Node{}, fmt.Errorf("error unmarshalling /common/nodes response %s", err)
	}

	return nodes, nil
}

func GetNodeByVMCID(c config.Cpi, cid string) (Node, error) {
	nodes, err := GetNodes(c)
	if err != nil {
		return Node{}, err
	}

	for _, node := range nodes {
		if node.CID == cid {
			return node, nil
		}
	}

	return Node{}, fmt.Errorf("vm with cid: %s was not found", cid)
}

func GetNodeByDiskCID(c config.Cpi, diskCid string) (Node, error) {
	nodes, err := GetNodes(c)
	if err != nil {
		return Node{}, err
	}

	for _, node := range nodes {
		if node.PersistentDisk.DiskCID == diskCid {
			return node, nil
		}
	}

	return Node{}, fmt.Errorf("vm with cid: %s was not found", diskCid)
}

func GetOBMSettings(c config.Cpi, nodeID string) ([]OBMSetting, error) {
	nodeURL := fmt.Sprintf("http://%s/api/common/nodes/%s", c.ApiServer, nodeID)
	resp, err := http.Get(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("error getting node %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed getting node with status: %s, err: %s", resp.Status, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading node body %s", err)
	}

	var node Node
	err = json.Unmarshal(b, &node)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal node body %s", err)
	}

	if len(node.OBMSettings) == 0 {
		return nil, errors.New("error: got empty obm settings")
	}

	return node.OBMSettings, nil
}

func IsAMTService(c config.Cpi, nodeID string) (bool, error) {
	obmSettings, err := GetOBMSettings(c, nodeID)
	if err != nil {
		return false, fmt.Errorf("error retrieving obm settings of node: %s, error: %v", nodeID, err)
	}

	for _, setting := range obmSettings {
		if setting.ServiceName == OBMSettingAMTServiceName {
			return true, nil
		}
	}

	return false, nil
}

func ReleaseNode(c config.Cpi, nodeID string) error {
	blockFlag := []byte(fmt.Sprintf("{\"status\": \"%s\"}", Available))
	return PatchNode(c, nodeID, blockFlag)
}

func GetNodeCatalog(c config.Cpi, nodeID string) (NodeCatalog, error) {
	catalogURL := fmt.Sprintf("http://%s/api/common/nodes/%s/catalogs/ohai", c.ApiServer, nodeID)
	resp, err := http.Get(catalogURL)
	if err != nil {
		return NodeCatalog{}, fmt.Errorf("error getting catalog %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return NodeCatalog{}, fmt.Errorf("Failed getting node catalog with status: %s, err: %s", resp.Status, err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return NodeCatalog{}, fmt.Errorf("error reading catalog body %s", err)
	}

	var nodeCatalog NodeCatalog
	err = json.Unmarshal(b, &nodeCatalog)
	if err != nil {
		return NodeCatalog{}, fmt.Errorf("error unmarshal catalog body %s", err)
	}

	return nodeCatalog, nil
}

func BlockNode(c config.Cpi, nodeID string) error {
	blockFlag := []byte(fmt.Sprintf("{\"status\": \"%s\", \"status_reason\": \"%s\"}", Blocked, DiskReason))
	return PatchNode(c, nodeID, blockFlag)
}

func SetNodeMetadata(c config.Cpi, nodeID string, metadata string) error {
	metadataBytes := []byte(fmt.Sprintf("{\"metadata\": %s}", metadata))
	return PatchNode(c, nodeID, metadataBytes)
}

func PatchNode(c config.Cpi, nodeID string, body []byte) error {
	url := fmt.Sprintf("http://%s/api/common/nodes/%s", c.ApiServer, nodeID)

	request, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("Error building request to api server: %s", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = int64(len(body))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error making request to api server: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed patching URL: %s with status: %s", url, resp.Status)
	}

	return nil
}
func MakeDiskRequest(c config.Cpi, node Node, newDiskState bool) error {

	container := PersistentDiskSettingsContainer{
		PersistentDisk: node.PersistentDisk,
	}
	container.PersistentDisk.IsAttached = newDiskState

	bodyBytes, err := json.Marshal(container)
	if err != nil {
		return err
	}

	err = PatchNode(c, node.ID, bodyBytes)
	if err != nil {
		return fmt.Errorf("Error requesting new disk state: %v", err)
	}

	return nil
}
