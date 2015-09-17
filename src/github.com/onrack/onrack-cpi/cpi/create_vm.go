package cpi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
	"github.com/onrack/onrack-cpi/workflows"
)

func CreateVM(c config.Cpi, extInput bosh.ExternalInput) (string, error) {
	agentID, stemcellCID, boshNetworks, err := parseCreateVMInput(extInput)
	if err != nil {
		return "", err
	}

	var netSpec bosh.Network
	var netName string
	for k, v := range boshNetworks {
		netName = k
		netSpec = v
	}

	nodes, err := onrackhttp.GetNodes(c)
	if err != nil {
		return "", err
	}

	nodeID, err := selectNonReservedNode(nodes)
	if err != nil {
		return "", err
	}

	if netSpec.NetworkType == bosh.ManualNetworkType {
		nodeCatalog, err := onrackhttp.GetNodeCatalog(c, nodeID)
		if err != nil {
			return "", err
		}

		netSpec, err = attachMAC(nodeCatalog.Data.NetworkData.Networks, netSpec)
		if err != nil {
			return "", fmt.Errorf("error attaching mac address %s", err)
		}
	}

	env := bosh.AgentEnv{
		AgentID:   agentID,
		Blobstore: c.Agent.Blobstore,
		Disks: map[string]string{
			"system":    "/dev/sda",
			"ephemeral": "/dev/sdb",
		},
		Mbus:     c.Agent.Mbus,
		Networks: map[string]bosh.Network{netName: netSpec},
		NTP:      c.Agent.Ntp,
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
	vmCID, err := onrackhttp.UploadFile(c, nodeID, envReader, int64(len(envBytes)))
	if err != nil {
		return "", err
	}

	agentRegistryName := fmt.Sprintf("agent-%s", vmCID)
	regBytes, err := json.Marshal(bosh.DefaultAgentRegistrySettings())
	if err != nil {
		return "", fmt.Errorf("error marshalling agent env %s", err)
	}
	regReader := bytes.NewReader(regBytes)
	regUUID, err := onrackhttp.UploadFile(c, agentRegistryName, regReader, int64(len(regBytes)))
	if err != nil {
		return "", err
	}

	log.Printf("Succeeded uploading agent registry, got '%s' as uuid", regUUID)
	uploadReq := workflows.UploadAgentSettingsRequest{
		Name: workflows.OnrackReserveVMGraphName,
		Options: map[string]workflows.UploadAgentSettingsOptions{
			"defaults": workflows.UploadAgentSettingsOptions{
				AgentSettingsFile:    nodeID,
				AgentSettingsPath:    workflows.OnrackEnvPath,
				CID:                  vmCID,
				RegistrySettingsFile: agentRegistryName,
				RegistrySettingsPath: workflows.OnrackRegistryPath,
				StemcellFile:         stemcellCID,
			},
		},
	}

	//refactor pending on workflow uploading
	createVMURL := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows", c.ApiServer, nodeID)
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

func attachMAC(nodeNetworks map[string]onrackhttp.Network, oldSpec bosh.Network) (bosh.Network, error) {
	var upNetworks []onrackhttp.Network

	for _, nodeNetwork := range nodeNetworks {
		if nodeNetwork.State == onrackhttp.NetworkActive && nodeNetwork.Encapsulation == onrackhttp.EthernetNetwork {
			upNetworks = append(upNetworks, nodeNetwork)
		}
	}

	if len(upNetworks) == 0 {
		log.Println("node has no active network")
		return bosh.Network{}, errors.New("node has no active network")
	}

	if len(upNetworks) > 1 {
		log.Printf("node has %d active networks", len(upNetworks))
		return bosh.Network{}, fmt.Errorf("node has %d active networks", len(upNetworks))
	}

	var nodeMac string
	for netName, netValue := range upNetworks[0].Addresses {
		if netValue.Family == onrackhttp.MacAddressFamily {
			nodeMac = netName
		}
	}

	net := bosh.Network{
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

func selectNonReservedNode(nodes []onrackhttp.Node) (string, error) {
	availableNodes := rejectReservedNodes(nodes)
	if len(availableNodes) == 0 {
		return "", errors.New("all nodes have been reserved")
	}

	t := time.Now()
	rand.Seed(t.Unix())

	i := rand.Intn(len(availableNodes))
	return availableNodes[i].ID, nil
}

func rejectReservedNodes(nodes []onrackhttp.Node) []onrackhttp.Node {
	var n []onrackhttp.Node

	for i := range nodes {
		if nodes[i].Reserved == "" {
			n = append(n, nodes[i])
		}
	}

	return n
}
