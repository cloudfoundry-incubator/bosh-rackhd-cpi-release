package onrackhttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/onrack/onrack-cpi/config"
)

func GetNodes(c config.Cpi) ([]Node, error) {
	nodesURL := fmt.Sprintf("http://%s:8080/api/common/nodes", c.ApiServer)
	resp, err := http.Get(nodesURL)
	if err != nil {
		log.Printf("error fetching nodes %s", err)
		return []Node{}, fmt.Errorf("error fetching nodes %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error getting nodes %s", err)
		return []Node{}, fmt.Errorf("Failed getting nodes with status: %s", resp.Status)
	}

	nodeBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading node response body %s", err)
		return []Node{}, fmt.Errorf("error reading node response body %s", err)
	}

	var nodes []Node
	err = json.Unmarshal(nodeBytes, &nodes)
	if err != nil {
		log.Printf("error unmarshalling /common/nodes response %s", err)
		return []Node{}, fmt.Errorf("error unmarshalling /common/nodes response %s", err)
	}

	return nodes, nil
}

func GetNodeCatalog(c config.Cpi, nodeID string) (NodeCatalog, error) {
	catalogURL := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/catalogs/ohai", c.ApiServer, nodeID)
	resp, err := http.Get(catalogURL)
	if err != nil {
		log.Printf("error getting catalog %s", err)
		return NodeCatalog{}, fmt.Errorf("error getting catalog %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error getting nodes %s", err)
		return NodeCatalog{}, fmt.Errorf("Failed getting nodes with status: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading catalog body %s", err)
		return NodeCatalog{}, fmt.Errorf("error reading catalog body %s", err)
	}

	var nodeCatalog NodeCatalog
	err = json.Unmarshal(b, &nodeCatalog)
	if err != nil {
		log.Printf("error unmarshal catalog body %s", err)
		return NodeCatalog{}, fmt.Errorf("error unmarshal catalog body %s", err)
	}

	return nodeCatalog, nil
}
