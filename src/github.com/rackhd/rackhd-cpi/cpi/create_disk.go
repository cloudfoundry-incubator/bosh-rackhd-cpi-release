package cpi

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/nu7hatch/gouuid"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

func CreateDisk(c config.Cpi, extInput bosh.MethodArguments) (string, error) {
	diskSizeInMB, vmCID, err := parseCreateDiskInput(extInput)
	if err != nil {
		return "", err
	}

	var node rackhdapi.Node
	if vmCID != "" {
		node, err = rackhdapi.GetNodeByVMCID(c, vmCID)
		if err != nil {
			return "", err
		}

		if node.PersistentDisk.DiskCID != "" {
			return "", fmt.Errorf("error creating disk: VM %s already has a persistent disk", vmCID)
		}
	} else {
		nodeID, err := SelectNodeFromRackHD(c, "")
		if err != nil {
			return "", err
		}
		node.ID = nodeID
	}

	catalog, err := rackhdapi.GetNodeCatalog(c, node.ID)
	if err != nil {
		return "", fmt.Errorf("error getting catalog of VM: %s", vmCID)
	}

	availableSpaceInKB, err := strconv.Atoi(catalog.Data.BlockDevices[rackhdapi.PersistentDiskLocation].Size)
	if err != nil {
		return "", fmt.Errorf("error creating disk for VM %s: disk not found", vmCID)
	}

	if availableSpaceInKB < diskSizeInMB*1024 {
		return "", fmt.Errorf("error creating disk with size %vMB for VM %s: insufficient available disk space", diskSizeInMB, vmCID)
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("error generating UUID: %s", err)
	}

	container := rackhdapi.PersistentDiskSettingsContainer{
		PersistentDisk: rackhdapi.PersistentDiskSettings{
			DiskCID:    uuid.String(),
			Location:   fmt.Sprintf("/dev/%s", rackhdapi.PersistentDiskLocation),
			IsAttached: false,
		},
	}

	bodyBytes, err := json.Marshal(container)
	if err != nil {
		return "", fmt.Errorf("error marshalling persistent disk information for VM %s", vmCID)
	}

	err = rackhdapi.PatchNode(c, node.ID, bodyBytes)
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}

func parseCreateDiskInput(extInput bosh.MethodArguments) (int, string, error) {
	diskSizeInput := extInput[0]
	diskSizeInMB := int(diskSizeInput.(float64))

	vmCIDInput := extInput[2]
	var vmCID string
	if reflect.TypeOf(vmCID) != reflect.TypeOf(vmCIDInput) {
		return 0, "", fmt.Errorf("vmCIDInput is unexpected type")
	}

	vmCID = vmCIDInput.(string)

	return diskSizeInMB, vmCID, nil
}
