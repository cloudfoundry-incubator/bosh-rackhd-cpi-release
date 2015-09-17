package cpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"time"

	"github.com/onrack/onrack-cpi/config"
)

const (
	onrackNetworkActive    = "up"
	onrackNetworkInactive  = "down"
	onrackEthernetNetwork  = "Ethernet"
	onrackMacAddressFamily = "lladdr"
)

const (
	onrackGraphName    = "Graph.CF.CreateReserveVM"
	onrackEnvPath      = "/var/vcap/bosh/onrack-cpi-agent-env.json"
	onrackRegistryPath = "/var/vcap/bosh/agent.json"
)

const (
	boshDynamicNetworkType = "dynamic"
	boshManualNetworkType  = "manual"
)

func CreateVM(config config.Cpi, extInput ExternalInput) (string, error) {
	// parse external input
	agentID, stemcellCID, boshNetworks, err := parseCreateVMInput(extInput)
	if err != nil {
		return "", err
	}

	var netSpec boshNetwork
	var netName string
	for k, v := range boshNetworks {
		netName = k
		netSpec = v
	}

	nodesURL := fmt.Sprintf("http://%s:8080/api/common/nodes", config.ApiServer)
	resp, err := http.Get(nodesURL)
	if err != nil {
		return "", fmt.Errorf("error fetching nodes %s", err)
	}
	defer resp.Body.Close()

	nodeBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	nodeID, err := selectNonReservedNode(nodeBytes)
	if err != nil {
		return "", err
	}

	if netSpec.NetworkType == boshManualNetworkType {
		catalogURL := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/catalogs/ohai", config.ApiServer, nodeID)
		resp, err := http.Get(catalogURL)
		if err != nil {
			return "", fmt.Errorf("error getting catalog %s", err)
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("error reading catalog body %s", err)
		}

		var nodeCatalog onrackCatalogResponse
		err = json.Unmarshal(b, &nodeCatalog)
		if err != nil {
			return "", fmt.Errorf("error unmarshal catalog body %s", err)
		}

		netSpec, err = attachMAC(nodeCatalog.Data.NetworkData.Networks, netSpec)
		if err != nil {
			return "", fmt.Errorf("error attaching mac address %s", err)
		}
	}

	env := agentEnv{
		AgentID:   agentID,
		Blobstore: config.Agent.Blobstore,
		Disks: map[string]string{
			"system":    "/dev/sda",
			"ephemeral": "/dev/sdb",
		},
		Mbus:     config.Agent.Mbus,
		Networks: map[string]boshNetwork{netName: netSpec},
		NTP:      config.Agent.Ntp,
		VM: map[string]string{
			"id":   nodeID,
			"name": nodeID,
		},
	}

	envBytes, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("error marshalling agent env %s", err)
	}
	envReader := bytes.NewReader(envBytes)
	vmCID, err := onrackhttp.UploadFile(config, nodeID, envReader, int64(len(envBytes)))
	if err != nil {
		return "", err
	}

	agentRegistryName := fmt.Sprintf("agent-%s", vmCID)
	regBytes, err := json.Marshal(bosh.DefaultAgentRegistrySettings())
	if err != nil {
		return "", fmt.Errorf("error marshalling agent env %s", err)
	}
	regReader := bytes.NewReader(regBytes)
	regUUID, err := onrackhttp.UploadFile(config, agentRegistryName, regReader, int64(len(regBytes)))
	if err != nil {
		return "", err
	}

	log.Printf("Succeeded uploading agent registry, got '%s' as uuid", regUUID)
	uploadReq := uploadAgentSettingsRequest{
		Name: onrackGraphName,
		Options: map[string]uploadAgentSettingsOptions{
			"defaults": uploadAgentSettingsOptions{
				AgentSettingsFile:    nodeID,
				AgentSettingsPath:    onrackEnvPath,
				CID:                  vmCID,
				RegistrySettingsFile: agentRegistryName,
				RegistrySettingsPath: onrackRegistryPath,
				StemcellFile:         stemcellCID,
			},
		},
	}

	createVMURL := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows", config.ApiServer, nodeID)
	createVMBytes, err := json.Marshal(uploadReq)
	if err != nil {
		return "", fmt.Errorf("error marshalling agent env %s", err)
	}
	createVMReader := bytes.NewReader(createVMBytes)
	createVMReq, err := http.NewRequest("POST", createVMURL, createVMReader)
	if err != nil {
		return "", fmt.Errorf("error creating request for create vm %s", err)
	}
	createVMReq.Header.Set("Content-Type", "application/json")
	createVMResp, err := http.DefaultClient.Do(createVMReq)
	if err != nil {
		return "", fmt.Errorf("error making request start create vm workflow %s", err)
	}

	if createVMResp.StatusCode != 201 {
		return "", fmt.Errorf("error creating vm workflow %s", err)
	}

	return vmCID, nil
}

func parseCreateVMInput(extInput ExternalInput) (string, string, map[string]boshNetwork, error) {
	networkSpecs := map[string]boshNetwork{}
	agentIDInput := extInput[0]
	var agentID string

	if reflect.TypeOf(agentIDInput) != reflect.TypeOf(agentID) {
		log.Printf("agent id has unexpected type: %s. Expecting a string", reflect.TypeOf(agentIDInput))
		return "", "", networkSpecs, fmt.Errorf("agent id has unexpected type: %s. Expecting a string", reflect.TypeOf(agentIDInput))
	}

	agentID = agentIDInput.(string)
	if agentID == "" {
		log.Println("agent id cannot be empty")
		return "", "", networkSpecs, errors.New("agent id cannot be empty")
	}

	stemcellIDInput := extInput[1]
	var stemcellID string

	if reflect.TypeOf(stemcellIDInput) != reflect.TypeOf(stemcellID) {
		log.Printf("stemcell id has unexpected type: %s. Expecting a string", reflect.TypeOf(stemcellIDInput))
		return "", "", networkSpecs, fmt.Errorf("stemcell id has unexpected type: %s. Expecting a string", reflect.TypeOf(stemcellIDInput))
	}

	stemcellID = stemcellIDInput.(string)
	if stemcellID == "" {
		log.Println("stemcell id cannot be empty")
		return "", "", networkSpecs, errors.New("stemcell id cannot be empty")
	}

	networkInput := extInput[3]
	var networks map[string]interface{}

	if reflect.TypeOf(networkInput) != reflect.TypeOf(networks) {
		log.Printf("network config has unexpected type: %s. Expecting a map", reflect.TypeOf(networkInput))
		return "", networkSpecs, fmt.Errorf("network config has unexpected type in: %s. Expecting a map", reflect.TypeOf(networkInput))
	}

	networks = networkInput.(map[string]interface{})
	if len(networks) > 1 {
		log.Printf("config error: Only one network supported %d", len(networks))
		return "", networkSpecs, fmt.Errorf("config error: Only one network supported, provided length: %d", len(networks))
	}

	b, err := json.Marshal(networks)
	if err != nil {
		panic(err)
	}
	var boshNetworks map[string]boshNetwork
	err = json.Unmarshal(b, &boshNetworks)
	if err != nil {
		panic(err)
	}

	var boshNetName string
	var boshNet boshNetwork

	for k, v := range boshNetworks {
		boshNetName = k
		boshNet = v
	}

	if valErr := validateNetworkingConfig(boshNet); valErr != nil {
		return "", networkSpecs, valErr
	}

	networkSpecs[boshNetName] = boshNetwork{
		NetworkType: boshNet.NetworkType,
		Netmask:     boshNet.Netmask,
		Gateway:     boshNet.Gateway,
		IP:          boshNet.IP,
		Default:     boshNet.Default,
		DNS:         boshNet.DNS,
	}

	return stemcellID, networkSpecs, nil
}

func validateNetworkingConfig(bn boshNetwork) error {
	if bn.NetworkType == boshManualNetworkType {
		return validateManualNetworkingConfig(bn)
	}

	if bn.NetworkType == boshDynamicNetworkType {
		return nil
	}

	return fmt.Errorf("unexpected or empty network type: %s\n", bn.NetworkType)
}

func validateManualNetworkingConfig(bn boshNetwork) error {
	if bn.IP == "" {
		log.Println("config error: ip must be specified for manual network")
		return errors.New("config error: ip must be specified for manual network")
	}

	if bn.Gateway == "" {
		log.Println("config error: gateway must be specified for manual network")
		return errors.New("config error: gateway must be specified for manual network")
	}

	if bn.Netmask == "" {
		log.Println("config error: netmask must be specified for manual network")
		return errors.New("config error: netmask must be specified for manual network")
	}

	return nil
}

func attachMAC(nodeNetworks map[string]onrackNetwork, oldSpec boshNetwork) (boshNetwork, error) {
	var upNetworks []onrackNetwork

	for _, nodeNetwork := range nodeNetworks {
		if nodeNetwork.State == onrackNetworkActive && nodeNetwork.Encapsulation == onrackEthernetNetwork {
			upNetworks = append(upNetworks, nodeNetwork)
		}
	}

	if len(upNetworks) == 0 {
		log.Println("node has no active network")
		return boshNetwork{}, errors.New("node has no active network")
	}

	if len(upNetworks) > 1 {
		log.Printf("node has %d active networks", len(upNetworks))
		return boshNetwork{}, fmt.Errorf("node has %d active networks", len(upNetworks))
	}

	var nodeMac string
	for netName, netValue := range upNetworks[0].Addresses {
		if netValue.Family == onrackMacAddressFamily {
			nodeMac = netName
		}
	}

	net := boshNetwork{
		NetworkType: oldSpec.NetworkType,
		Netmask:     oldSpec.Netmask,
		Gateway:     oldSpec.Gateway,
		IP:          oldSpec.IP,
		Default:     oldSpec.Default,
		DNS:         oldSpec.DNS,
		MAC:         nodeMac,
	}

	return net, nil
}

func selectNonReservedNode(onRackNode []byte) (string, error) {
	var nodes []node
	err := json.Unmarshal(onRackNode, &nodes)
	if err != nil {
		log.Printf("Error unmarshalling /common/nodes response %s", err)
		return "", fmt.Errorf("Error unmarshalling /common/nodes response %s", err)
	}

	availableNodes := rejectReservedNodes(nodes)
	if len(availableNodes) == 0 {
		return "", errors.New("all nodes have been reserved")
	}

	t := time.Now()
	rand.Seed(t.Unix())

	i := rand.Intn(len(availableNodes))
	return availableNodes[i].ID, nil
}

func rejectReservedNodes(nodes []node) []node {
	var n []node

	for i := range nodes {
		if nodes[i].Reserved == "" {
			n = append(n, nodes[i])
		}
	}

	return n
}

type uploadAgentSettingsRequest struct {
	Name    string                                `json:"name"`
	Options map[string]uploadAgentSettingsOptions `json:"options"`
}

type uploadAgentSettingsOptions struct {
	AgentSettingsFile    string `json:"agentSettingsFile"`
	AgentSettingsPath    string `json:"agentSettingsPath"`
	CID                  string `json:"cid"`
	DownloadDir          string `json:"downloadDir"`
	RegistrySettingsFile string `json:"registrySettingsFile"`
	RegistrySettingsPath string `json:"registrySettingsPath"`
	StemcellFile         string `json:"stemcellFile"`
}

type node struct {
	Workflows []interface{} `json:"workflows"`
	Reserved  string        `json:"reserved"`
	ID        string        `json:"id"`
}

type boshNetwork struct {
	NetworkType     string                 `json:"type"`
	Netmask         string                 `json:"netmask"`
	Gateway         string                 `json:"gateway"`
	IP              string                 `json:"ip"`
	Default         []string               `json:"default"`
	DNS             []string               `json:"dns,omitempty"`
	CloudProperties map[string]interface{} `json:"cloud_properties"`
	MAC             string                 `json:"mac"`
}

type onrackCatalogResponse struct {
	Data onrackCatalogData `json:"data"`
}

type onrackCatalogData struct {
	NetworkData onRackNetworkCatalog `json:"network"`
}

type onRackNetworkCatalog struct {
	Networks map[string]onrackNetwork `json:"interfaces"`
}

type onrackNetwork struct {
	Encapsulation string                          `json:"encapsulation"`
	Number        string                          `json:"number"`
	Addresses     map[string]onrackNetworkAddress `json:"addresses"`
	State         string                          `json:"state"`
}

type onrackNetworkAddress struct {
	Family string `json:"family"`
}

type agentEnv struct {
	AgentID   string                 `json:"agent_id"`
	Blobstore map[string]interface{} `json:"blobstore"`
	Disks     map[string]string      `json:"disks"`
	Env       map[string]interface{} `json:"env"`
	Mbus      string                 `json:"mbus"`
	Networks  map[string]boshNetwork `json:"networks"`
	NTP       []string               `json:"ntp"`
	VM        map[string]string      `json:"vm"`
}
