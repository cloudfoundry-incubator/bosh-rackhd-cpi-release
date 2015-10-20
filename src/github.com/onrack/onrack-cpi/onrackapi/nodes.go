package onrackapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/onrack/onrack-cpi/config"
)

const (
	NetworkActive    = "up"
	NetworkInactive  = "down"
	EthernetNetwork  = "Ethernet"
	MacAddressFamily = "lladdr"
)

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

func GetNodes(c config.Cpi) ([]Node, error) {
	nodesURL := fmt.Sprintf("http://%s:8080/api/common/nodes", c.ApiServer)
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

func ReleaseNode(c config.Cpi, nodeID string) error {
	url := fmt.Sprintf("http://%s:8080/api/common/nodes/%s", c.ApiServer, nodeID)
	reserveFlag := `{"reserved" : ""}`
	body := ioutil.NopCloser(strings.NewReader(reserveFlag))
	defer body.Close()

	request, err := http.NewRequest("PATCH", url, body)
	if err != nil {
		return fmt.Errorf("Error building request to api server: %s", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = int64(len(reserveFlag))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error making request to api server: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed patching with status: %s", resp.Status)
	}

	return nil
}

func GetNodeCatalog(c config.Cpi, nodeID string) (NodeCatalog, error) {
	catalogURL := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/catalogs/ohai", c.ApiServer, nodeID)
	resp, err := http.Get(catalogURL)
	if err != nil {
		return NodeCatalog{}, fmt.Errorf("error getting catalog %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return NodeCatalog{}, fmt.Errorf("Failed getting nodes with status: %s, err: %s", resp.Status, err)
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
