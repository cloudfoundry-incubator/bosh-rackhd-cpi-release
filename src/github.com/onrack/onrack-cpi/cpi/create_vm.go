package cpi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
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

	defer onrackapi.ReleaseNode(c, nodeID)

	var netSpec bosh.Network
	var netName string
	for k, v := range boshNetworks {
		netName = k
		netSpec = v
	}

	if netSpec.NetworkType == bosh.ManualNetworkType {
		nodeCatalog, err := onrackapi.GetNodeCatalog(c, nodeID)
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
	vmCID, err := onrackapi.UploadFile(c, nodeID, envReader, int64(len(envBytes)))
	if err != nil {
		return "", err
	}
	defer onrackapi.DeleteFile(c, nodeID)

	agentRegistryName := fmt.Sprintf("agent-%s", vmCID)
	regBytes, err := json.Marshal(bosh.DefaultAgentRegistrySettings())
	if err != nil {
		return "", fmt.Errorf("error marshalling agent env %s", err)
	}
	regReader := bytes.NewReader(regBytes)
	regUUID, err := onrackapi.UploadFile(c, agentRegistryName, regReader, int64(len(regBytes)))
	if err != nil {
		return "", err
	}
	defer onrackapi.DeleteFile(c, agentRegistryName)
	log.Info(fmt.Sprintf("Succeeded uploading agent registry, got '%s' as uuid", regUUID))

	publicKeyName := fmt.Sprintf("key-%s", vmCID)
	publicKeyReader := strings.NewReader(publicKey)
	keyUUID, err := onrackapi.UploadFile(c, publicKeyName, publicKeyReader, int64(len(publicKey)))
	if err != nil {
		return "", err
	}
	defer onrackapi.DeleteFile(c, publicKeyName)
	log.Info(fmt.Sprintf("Succeeded uploading public key, got '%s' as uuid", keyUUID))

	workflowName, err := workflows.PublishProvisionNodeWorkflow(c, vmCID)
	if err != nil {
		log.Error(fmt.Sprintf("error publishing provision workflow: %s", err))
		return "", fmt.Errorf("error publishing provision workflow: %s", err)
	}

	envPath := onrackapi.OnrackEnvPath
	regPath := onrackapi.OnrackRegistryPath
	options := workflows.ProvisionNodeWorkflowOptions{
		AgentSettingsFile:    &nodeID,
		AgentSettingsPath:    &envPath,
		CID:                  &vmCID,
		PublicKeyFile:        &publicKeyName,
		RegistrySettingsFile: &agentRegistryName,
		RegistrySettingsPath: &regPath,
		StemcellFile:         &stemcellCID,
	}

	err = workflows.RunProvisionNodeWorkflow(c, nodeID, workflowName, options)
	if err != nil {
		log.Error(fmt.Sprintf("error running provision workflow: %s", err))
		return "", fmt.Errorf("error running provision workflow: %s", err)
	}

	return vmCID, nil
}

func attachMAC(nodeNetworks map[string]onrackapi.Network, oldSpec bosh.Network) (bosh.Network, error) {
	var upNetworks []onrackapi.Network

	for _, nodeNetwork := range nodeNetworks {
		if nodeNetwork.State == onrackapi.NetworkActive && nodeNetwork.Encapsulation == onrackapi.EthernetNetwork {
			upNetworks = append(upNetworks, nodeNetwork)
		}
	}

	if len(upNetworks) == 0 {
		log.Error("node has no active network")
		return bosh.Network{}, errors.New("node has no active network")
	}

	if len(upNetworks) > 1 {
		log.Error(fmt.Sprintf("node has %d active networks", len(upNetworks)))
		return bosh.Network{}, fmt.Errorf("node has %d active networks", len(upNetworks))
	}

	var nodeMac string
	for netName, netValue := range upNetworks[0].Addresses {
		if netValue.Family == onrackapi.MacAddressFamily {
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
			log.Error(fmt.Sprintf("retry %d: error choosing node %s", i, err))
			continue
		}

		reserved, err = reserve(c, nodeID)
		if err != nil {
			log.Error(fmt.Sprintf("retry %d: error reserving node %s", i, err))
			err = onrackapi.ReleaseNode(c, nodeID)
			if err != nil {
				log.Error(fmt.Sprintf("error releasing node %s, %s", nodeID, err))
			}
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
	u, err := uuid.NewV4()
	if err != nil {
		return "", errors.New("error generating UUID")
	}
	uStr := u.String()

	workflowName, err := workflows.PublishReserveNodeWorkflow(c, u.String())
	if err != nil {
		log.Error(fmt.Sprintf("error publishing reserve workflow: %s", err))
		return "", fmt.Errorf("error publishing reserve workflow: %s", err)
	}

	o := workflows.ReserveNodeWorkflowOptions{UUID: &uStr}
	err = workflows.RunReserveNodeWorkflow(c, nodeID, workflowName, o)
	if err != nil {
		log.Error(fmt.Sprintf("error running reserve workflow: %s", err))
		return "", fmt.Errorf("error running reserve workflow: %s", err)
	}

	log.Info(fmt.Sprintf("reserved node %s", nodeID))
	return nodeID, nil
}

func selectNodeFromOnRack(c config.Cpi) (string, error) {
	nodes, err := onrackapi.GetNodes(c)
	if err != nil {
		return "", err
	}

	nodeID, err := selectNonReservedNode(nodes)
	if err != nil {
		return "", err
	}

	log.Info(fmt.Sprintf("selected node %s", nodeID))
	return nodeID, nil
}

func selectNonReservedNode(nodes []onrackapi.Node) (string, error) {
	availableNodes := rejectReservedNodes(nodes)
	if len(availableNodes) == 0 {
		return "", errors.New("all nodes have been reserved")
	}

	t := time.Now()
	rand.Seed(t.Unix())

	i := rand.Intn(len(availableNodes))
	return availableNodes[i].ID, nil
}

func rejectReservedNodes(nodes []onrackapi.Node) []onrackapi.Node {
	var n []onrackapi.Node

	for i := range nodes {
		if nodes[i].Reserved == "" && nodes[i].CID == "" {
			n = append(n, nodes[i])
		}
	}

	return n
}
