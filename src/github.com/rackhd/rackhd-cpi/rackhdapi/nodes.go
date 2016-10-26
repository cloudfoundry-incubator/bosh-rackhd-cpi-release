package rackhdapi

import (
  "bytes"
  "encoding/json"
  "errors"
  "fmt"
  "io/ioutil"
  "net/http"

  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/helpers"
  "github.com/rackhd/rackhd-cpi/models"
)

// GetNodes returns all nodes
func GetNodes(c config.Cpi) ([]models.Node, error) {
  url := fmt.Sprintf("%s/api/2.0/nodes", c.ApiServer)
  respBody, err := helpers.MakeRequest(url, "GET", 200, nil)
  if err != nil {
    return []models.Node{}, fmt.Errorf("error getting nodes: %s", err)
  }

  var nodes []models.Node
  err = json.Unmarshal(respBody, &nodes)
  if err != nil {
    return []models.Node{}, fmt.Errorf("error unmarshalling all nodes: %s", err)
  }

  return nodes, nil
}

func GetNode(c config.Cpi, nodeID string) (models.Node, error) {
  nodeURL := fmt.Sprintf("%s/api/2.0/nodes/%s", c.ApiServer, nodeID)
  resp, err := http.Get(nodeURL)
  if err != nil {
    return models.Node{}, fmt.Errorf("error fetching node %s: %s", nodeID, err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return models.Node{}, fmt.Errorf("Failed getting node %s with status: %s", nodeID, resp.Status)
  }

  nodeBytes, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return models.Node{}, fmt.Errorf("error reading node %s response body %s", nodeID, err)
  }

  var node models.Node
  err = json.Unmarshal(nodeBytes, &node)
  if err != nil {
    return models.Node{}, fmt.Errorf("error unmarshalling /2.0/node/%s response %s", nodeID, err)
  }

  return node, nil
}

func GetOBMSettings(c config.Cpi, nodeID string) ([]models.OBM, error) {
  nodeURL := fmt.Sprintf("%s/api/2.0/nodes/%s", c.ApiServer, nodeID)
  resp, err := http.Get(nodeURL)
  if err != nil {
    return nil, fmt.Errorf("error getting node %s", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return nil, fmt.Errorf("Failed getting node with status: %s, err: %s", resp.Status, err)
  }

  b, err := ioutil.ReadAll(resp.Body)
  fmt.Printf("obm body: %+v", string(b))
  if err != nil {
    return nil, fmt.Errorf("error reading node body %s", err)
  }

  var node models.Node
  err = json.Unmarshal(b, &node)
  if err != nil {
    return nil, fmt.Errorf("error unmarshal node body %s", err)
  }

  if len(node.OBMS) == 0 {
    return nil, errors.New("error: got empty obm settings")
  }

  return node.OBMS, nil
}

func GetOBMServiceName(c config.Cpi, nodeID string) (string, error) {
  obmSettings, err := GetOBMSettings(c, nodeID)
  if err != nil {
    return "", fmt.Errorf("error retrieving obm settings of node: %s, error: %v", nodeID, err)
  }

  return obmSettings[0].ServiceName, nil
}

func GetNodeCatalog(c config.Cpi, nodeID string) (models.NodeCatalog, error) {
  catalogURL := fmt.Sprintf("%s/api/2.0/nodes/%s/catalogs/ohai", c.ApiServer, nodeID)
  resp, err := http.Get(catalogURL)
  if err != nil {
    return models.NodeCatalog{}, fmt.Errorf("error getting catalog %s", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return models.NodeCatalog{}, fmt.Errorf("Failed getting node catalog with status: %s, err: %s", resp.Status, err)
  }

  b, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return models.NodeCatalog{}, fmt.Errorf("error reading catalog body %s", err)
  }

  var nodeCatalog models.NodeCatalog
  err = json.Unmarshal(b, &nodeCatalog)
  if err != nil {
    return models.NodeCatalog{}, fmt.Errorf("error unmarshal catalog body %s", err)
  }

  return nodeCatalog, nil
}

func SetNodeMetadata(c config.Cpi, nodeID string, metadata string) error {
  metadataBytes := []byte(fmt.Sprintf("{\"metadata\": %s}", metadata))
  return PatchNode(c, nodeID, metadataBytes)
}

func PatchNode(c config.Cpi, nodeID string, body []byte) error {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s", c.ApiServer, nodeID)

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

func MakeDiskRequest(c config.Cpi, node models.Node, newDiskState bool) error {
  container := models.PersistentDiskSettingsContainer{
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
