package cpi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
	"github.com/onrack/onrack-cpi/workflows"
)

func CreateVM(c config.Cpi, extInput bosh.MethodArguments) (string, error) {
	agentID, stemcellCID, publicKey, boshNetworks, err := parseCreateVMInput(extInput)
	if err != nil {
		return "", err
	}

	nodeID, err := tryReservation(c, selectNodeFromOnRack, reserveNodeFromOnRack)
	if err != nil {
		return "", err
	}

	defer onrackhttp.ReleaseNode(c, nodeID)

	var netSpec bosh.Network
	var netName string
	for k, v := range boshNetworks {
		netName = k
		netSpec = v
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
	defer onrackhttp.DeleteFile(c, nodeID)

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
	defer onrackhttp.DeleteFile(c, agentRegistryName)
	log.Printf("Succeeded uploading agent registry, got '%s' as uuid", regUUID)

	publicKeyName := fmt.Sprintf("key-%s", vmCID)
	publicKeyReader := strings.NewReader(publicKey)
	keyUUID, err := onrackhttp.UploadFile(c, publicKeyName, publicKeyReader, int64(len(publicKey)))
	if err != nil {
		return "", err
	}
	defer onrackhttp.DeleteFile(c, publicKeyName)
	log.Printf("Succeeded uploading public key, got '%s' as uuid", keyUUID)

	createVMReq := onrackhttp.RunWorkflowRequestBody{
		Name: onrackhttp.OnrackCreateVMGraphName,
		Options: map[string]interface{}{
			"defaults": workflows.UploadAgentSettingsOptions{
				AgentSettingsFile:    nodeID,
				AgentSettingsPath:    onrackhttp.OnrackEnvPath,
				CID:                  vmCID,
				PublicKeyFile:        publicKeyName,
				RegistrySettingsFile: agentRegistryName,
				RegistrySettingsPath: onrackhttp.OnrackRegistryPath,
				StemcellFile:         stemcellCID,
			},
		},
	}

	err = onrackhttp.RunWorkflow(c, nodeID, createVMReq)
	if err != nil {
		return "", err
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
		MAC:         strings.ToLower(nodeMac),
	}

	return net, nil
}

type selectionFunc func(config.Cpi) (string, error)
type reservationFunc func(config.Cpi, string) (string, error)

func tryReservation(c config.Cpi, choose selectionFunc, reserve reservationFunc) (string, error) {
	var reserved string
	for i := 0; i < c.MaxCreateVMAttempt; i++ {
		nodeID, err := choose(c)
		if err != nil {
			log.Printf("retry %d: error choosing node %s", i, err)
			continue
		}

		reserved, err = reserve(c, nodeID)
		if err != nil {
			log.Printf("retry %d: error reserving node %s", i, err)
			continue
		}

		if reserved != "" {
			break
		}
	}

	if reserved == "" {
		return "", errors.New("unable to reserve node")
	}

	return reserved, nil
}

func reserveNodeFromOnRack(c config.Cpi, nodeID string) (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", errors.New("error generating UUID")
	}

	workflowReq := onrackhttp.RunWorkflowRequestBody{
		Name: onrackhttp.OnrackReserveVMGraphName,
		Options: map[string]interface{}{
			"defaults": workflows.ReserveVMOptions{
				UUID: uuid.String(),
			},
		},
	}

	err = onrackhttp.RunWorkflow(c, nodeID, workflowReq)
	if err != nil {
		return "", fmt.Errorf("error reserving node %s", err)
	}

	log.Printf("reserved node %s", nodeID)
	return nodeID, nil
}

func selectNodeFromOnRack(c config.Cpi) (string, error) {
	nodes, err := onrackhttp.GetNodes(c)
	if err != nil {
		return "", err
	}

	nodeID, err := selectNonReservedNode(nodes)
	if err != nil {
		return "", err
	}

	log.Printf("selected node %s", nodeID)
	return nodeID, nil
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
