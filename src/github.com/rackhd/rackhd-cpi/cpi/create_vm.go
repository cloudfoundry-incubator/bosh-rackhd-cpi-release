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

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
	"github.com/rackhd/rackhd-cpi/workflows"
)

func CreateVM(c config.Cpi, extInput bosh.MethodArguments) (string, error) {
	agentID, stemcellCID, publicKey, boshNetworks, err := parseCreateVMInput(extInput)
	if err != nil {
		return "", err
	}

	nodeID, err := tryReservation(c, agentID, func(config.Cpi) error { return nil }, selectNodeFromRackHD, reserveNodeFromRackHD)
	if err != nil {
		return "", err
	}

	var netSpec bosh.Network
	var netName string
	for k, v := range boshNetworks {
		netName = k
		netSpec = v
	}

	if netSpec.NetworkType == bosh.ManualNetworkType {
		nodeCatalog, err := rackhdapi.GetNodeCatalog(c, nodeID)
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
		PublicKey: publicKey,
	}

	envBytes, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("error marshalling agent env %s", err)
	}
	envReader := bytes.NewReader(envBytes)
	vmCID, err := rackhdapi.UploadFile(c, nodeID, envReader, int64(len(envBytes)))
	if err != nil {
		return "", err
	}
	defer rackhdapi.DeleteFile(c, nodeID)

	workflowName, err := workflows.PublishProvisionNodeWorkflow(c, vmCID)
	if err != nil {
		return "", fmt.Errorf("error publishing provision workflow: %s", err)
	}

	err = workflows.RunProvisionNodeWorkflow(c, nodeID, workflowName, vmCID, stemcellCID)
	if err != nil {
		return "", fmt.Errorf("error running provision workflow: %s", err)
	}

	return vmCID, nil
}

func attachMAC(nodeNetworks map[string]rackhdapi.Network, oldSpec bosh.Network) (bosh.Network, error) {
	var upNetworks []rackhdapi.Network

	for _, nodeNetwork := range nodeNetworks {
		if nodeNetwork.State == rackhdapi.NetworkActive && nodeNetwork.Encapsulation == rackhdapi.EthernetNetwork {
			upNetworks = append(upNetworks, nodeNetwork)
		}
	}

	if len(upNetworks) == 0 {
		return bosh.Network{}, errors.New("node has no active network")
	}

	if len(upNetworks) > 1 {
		return bosh.Network{}, fmt.Errorf("node has %d active networks", len(upNetworks))
	}

	var nodeMac string
	for netName, netValue := range upNetworks[0].Addresses {
		if netValue.Family == rackhdapi.MacAddressFamily {
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

type filterFunc func(config.Cpi) error
type selectionFunc func(config.Cpi) (string, error)
type reservationFunc func(config.Cpi, string, string) error

func tryReservation(c config.Cpi, agentID string, filter filterFunc, choose selectionFunc, reserve reservationFunc) (string, error) {
	var nodeID string
	var err error
	for i := 0; i < c.MaxCreateVMAttempt; i++ {
		err = filter(c)
		if err != nil {
			log.Error(fmt.Sprintf("retry %d: error filtering nodes: %s", i, err))
			continue
		}

		nodeID, err = choose(c)

		if err != nil {
			log.Error(fmt.Sprintf("retry %d: error choosing node %s", i, err))
			continue
		}

		err = reserve(c, agentID, nodeID)
		if err != nil {
			log.Error(fmt.Sprintf("retry %d: error reserving node %s", i, err))
			defer rackhdapi.ReleaseNode(c, nodeID)
			continue
		}

		break
	}

	if err != nil {
		return "", errors.New("unable to reserve node")
	}

	return nodeID, nil
}

func blockNodesWithoutEphemeralDisk(c config.Cpi) error {
	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return err
	}

	for i := range nodes {
		if nodeIsAvailable(c, nodes[i]) {
			nodeCatalog, err := rackhdapi.GetNodeCatalog(c, nodes[i].ID)
			if err != nil {
				return err
			}
			if _, ok := nodeCatalog.Data.BlockDevices["sdb"]; !ok {
				rackhdapi.BlockNode(c, nodes[i].ID)
			}
		}
	}
	return nil
}

func reserveNodeFromRackHD(c config.Cpi, agentID string, nodeID string) error {
	workflowName, err := workflows.PublishReserveNodeWorkflow(c, agentID)
	if err != nil {
		return fmt.Errorf("error publishing reserve workflow: %s", err)
	}

	err = workflows.RunReserveNodeWorkflow(c, nodeID, workflowName)
	if err != nil {
		return fmt.Errorf("error running reserve workflow: %s", err)
	}

	log.Info(fmt.Sprintf("reserved node %s", nodeID))
	return nil
}

func selectNodeFromRackHD(c config.Cpi) (string, error) {
	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return "", err
	}

	nodeID, err := randomSelectAvailableNode(c, nodes)
	if err != nil || nodeID == "" {
		return "", err
	}

	log.Info(fmt.Sprintf("selected node %s", nodeID))
	return nodeID, nil
}

func randomSelectAvailableNode(c config.Cpi, nodes []rackhdapi.Node) (string, error) {
	availableNodes := getAllAvailableNodes(c, nodes)
	if len(availableNodes) == 0 {
		return "", errors.New("all nodes have been reserved")
	}

	t := time.Now()
	rand.Seed(t.Unix())

	i := rand.Intn(len(availableNodes))
	return availableNodes[i].ID, nil
}

func getAllAvailableNodes(c config.Cpi, nodes []rackhdapi.Node) []rackhdapi.Node {
	var n []rackhdapi.Node

	for i := range nodes {
		if nodeIsAvailable(c, nodes[i]) {
			n = append(n, nodes[i])
			log.Debug(fmt.Sprintf("node: %s is avaliable", nodes[i].ID))
		}
	}

	return n
}

func nodeIsAvailable(c config.Cpi, n rackhdapi.Node) bool {
	workflows, _ := rackhdapi.GetActiveWorkflows(c, n.ID)
	obmSettings, _ := rackhdapi.GetOBMSettings(c, n.ID)
	return (n.Status == "" || n.Status == rackhdapi.Available) && (n.CID == "") && (len(workflows) == 0) && (len(obmSettings) > 0)
}
