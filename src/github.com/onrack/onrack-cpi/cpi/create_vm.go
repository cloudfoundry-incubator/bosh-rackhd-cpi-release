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
	boshDynamicNetworkType = "dynamic"
	boshManualNetworkType  = "manual"
)

func CreateVM(config config.Cpi, extInput ExternalInput) (string, error) {
	// parse external input
	// selectNonReservedNode
	// buildNetworkConfig/attachMAC for manual
	return "", nil
}

func parseCreateVMInput(extInput ExternalInput) (string, map[string]boshNetwork, error) {
	networkSpecs := map[string]boshNetwork{}

	stemcellIDInput := extInput[1]
	var stemcellID string

	if reflect.TypeOf(stemcellIDInput) != reflect.TypeOf(stemcellID) {
		log.Printf("stemcell id has unexpected type: %s. Expecting a string", reflect.TypeOf(stemcellIDInput))
		return "", networkSpecs, fmt.Errorf("stemcell id has unexpected type: %s. Expecting a string", reflect.TypeOf(stemcellIDInput))
	}

	stemcellID = stemcellIDInput.(string)
	if stemcellID == "" {
		log.Println("stemcell id cannot be empty")
		return "", networkSpecs, errors.New("stemcell id cannot be empty")
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
