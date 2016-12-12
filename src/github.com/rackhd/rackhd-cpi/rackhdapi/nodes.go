package rackhdapi

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"

  log "github.com/Sirupsen/logrus"

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

// GetNodesWithType returns all nodes with the type specified in the query string
func GetNodesWithType(c config.Cpi, nodeType string) ([]models.Node, error) {
  url := fmt.Sprintf("%s/api/2.0/nodes?type=%s", c.ApiServer, nodeType)
  respBody, err := helpers.MakeRequest(url, "GET", 200, nil)
  if err != nil {
    return []models.Node{}, fmt.Errorf("error getting nodes: %s", err)
  }

  var nodes []models.Node
  err = json.Unmarshal(respBody, &nodes)
  if err != nil {
    return []models.Node{}, fmt.Errorf("error unmarshalling all type nodes: %s", err)
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
  if err != nil {
    return nil, fmt.Errorf("error reading node body %s", err)
  }

  var node models.Node
  err = json.Unmarshal(b, &node)
  if err != nil {
    return nil, fmt.Errorf("error unmarshal node body %s", err)
  }

  if len(node.OBMS) == 0 {
    return nil, fmt.Errorf("error: got empty obm settings")
  }

  return node.OBMS, nil
}

// GetOBMServiceName creates an OBMService if one is not assigned to the node
func GetOBMServiceName(c config.Cpi, nodeID string) (string, error) {
  obmSettings, err := GetOBMSettings(c, nodeID)
  if err != nil {
    if err.Error() == "error: got empty obm settings" {
      // create OBM Service, associate with a node
      name, er := setOBMService(c, nodeID)
      if er != nil {
        return "", err
      }
      return name, nil
    }
    return "", fmt.Errorf("error retrieving obm settings of node: %s, error: %v", nodeID, err)
  }

  return obmSettings[0].ServiceName, nil
}

// GetNodeCatalog returns a NodeCatalog object containing the full catalog for a given nodes' data
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

  _, err := helpers.MakeRequest(url, "PATCH", 200, body)
  if err != nil {
    return fmt.Errorf("Error making request to patch metadata to node: %s", err)
  }
  return nil
}

func MakeDiskRequest(c config.Cpi, node models.TagNode, newDiskState bool) error {
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

func setOBMService(c config.Cpi, nodeID string) (string, error) {
  log.Debug("Setting OBM Service from Environment")
  username := os.Getenv("OBM_USERNAME")
  if username == "" {
    log.Debug("OBM_USERNAME not found in environment, setting to default admin")
    username = "admin"
  }
  password := os.Getenv("OBM_PASSWORD")
  if password == "" {
    log.Debug("OBM_PASSWORD not found in environment, setting to default admin")
    password = "admin"
  }
  servicename := os.Getenv("OBM_SERVICE_NAME")
  if servicename == "" {
    log.Debug("OBM_SERVICE_NAME not found in enviornment, setting to default ipmi-obm-service")
    servicename = "ipmi-obm-service"
  }

  enclosureMAC, err := getEnclosureMACAddress(c, nodeID)
  if err != nil {
    return "", err
  }

  obmReq := &models.OBMServiceRequest{
    Config: models.OBMConfig{
      Host:     enclosureMAC,
      Password: password,
      User:     username,
    },
    NodeID:      nodeID,
    ServiceName: servicename,
  }

  url := fmt.Sprintf("%s/api/2.0/obms", c.ApiServer)
  log.Debug("Posting To %s with %+v", url, obmReq)
  obmBytes, err := json.Marshal(obmReq)
  if err != nil {
    return "", err
  }
  _, err = helpers.MakeRequest(url, "PUT", 201, obmBytes)
  if err != nil {
    return "", err
  }
  return servicename, nil
}

func getEnclosureMACAddress(c config.Cpi, nodeID string) (string, error) {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/catalogs/bmc", c.ApiServer, nodeID)
  resp, err := http.Get(url)
  if err != nil {
    return "", err
  }
  var catalogResp models.BMCCatalog
  respBytes, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return "", err
  }
  defer resp.Body.Close()

  err = json.Unmarshal(respBytes, &catalogResp)
  if err != nil {
    return "", err
  }
  if catalogResp.Data.MACAddress == "" {
    return "", fmt.Errorf("MAC Address not found")
  }

  return catalogResp.Data.MACAddress, nil
}
