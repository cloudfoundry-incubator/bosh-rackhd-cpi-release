package cpi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
	"github.com/rackhd/rackhd-cpi/workflows"
)

func CreateVM(c config.Cpi, extInput bosh.MethodArguments) (string, error) {
	agentID, stemcellCID, publicKey, boshNetworks, diskCID, err := parseCreateVMInput(extInput)
	if err != nil {
		return "", err
	}

	nodeID, err := TryReservation(c, diskCID, SelectNodeFromRackHD, ReserveNodeFromRackHD)
	if err != nil {
		return "", err
	}

	var netSpec bosh.Network
	var netName string
	for k, v := range boshNetworks {
		netName = k
		netSpec = v
	}

	nodeCatalog, err := rackhdapi.GetNodeCatalog(c, nodeID)
	if err != nil {
		return "", err
	}

	if netSpec.NetworkType == bosh.ManualNetworkType {
		netSpec, err = attachMAC(nodeCatalog.Data.NetworkData.Networks, netSpec)
		if err != nil {
			return "", err
		}
	}

	persistentMetadata := map[string]interface{}{}
	if _, sdbFound := nodeCatalog.Data.BlockDevices["sdb"]; sdbFound {
		persistentMetadata = map[string]interface{}{
			nodeID: map[string]string{
				"path": "/dev/sdb",
			},
		}
	}

	env := bosh.AgentEnv{
		AgentID:   agentID,
		Blobstore: c.Agent.Blobstore,
		Disks: map[string]interface{}{
			"system":     "/dev/sda",
			"persistent": persistentMetadata,
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

	workflowName, err := workflows.PublishProvisionNodeWorkflow(c)
	if err != nil {
		return "", fmt.Errorf("error publishing provision workflow: %s", err)
	}

	wipeDisk := (diskCID == "")

	err = workflows.RunProvisionNodeWorkflow(c, nodeID, workflowName, vmCID, stemcellCID, wipeDisk)
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
		return bosh.Network{}, errors.New("error attaching MAC address: node has no active network")
	}

	if len(upNetworks) > 1 {
		return bosh.Network{}, fmt.Errorf("error attaching MAC address: node has %d active networks", len(upNetworks))
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
