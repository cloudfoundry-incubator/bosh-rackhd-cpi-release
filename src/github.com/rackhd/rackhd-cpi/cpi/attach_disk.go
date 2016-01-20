package cpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

func AttachDisk(c config.Cpi, extInput bosh.MethodArguments) error {
	var vmCID string
	var diskCID string

	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(vmCID) {
		return errors.New("Received unexpected type for vm cid")
	}

	if reflect.TypeOf(extInput[1]) != reflect.TypeOf(diskCID) {
		return errors.New("Received unexpected type for disk cid")
	}

	vmCID = extInput[0].(string)
	diskCID = extInput[1].(string)

	node, err := rackhdapi.GetNodeByVMCID(c, vmCID)
	if err != nil {
		return fmt.Errorf("VM: %s not found\n", vmCID)
	}

	if node.PersistentDisk.DiskCID == "" {
		return fmt.Errorf("Disk: %s not found on VM: %s", diskCID, vmCID)
	}

	if node.PersistentDisk.DiskCID != diskCID {
		if node.PersistentDisk.IsAttached {
			return fmt.Errorf("Node %s has persistent disk %s attached. Cannot attach additional disk %s.", vmCID, node.PersistentDisk.DiskCID, diskCID)
		} else {
			return fmt.Errorf("Node %s has persistent disk %s, but detached. Cannot attach disk %s.", vmCID, node.PersistentDisk.DiskCID, diskCID)
		}
	}

	if node.CID != vmCID {
		return fmt.Errorf("Disk: %s does not belong to VM: %s\n", diskCID, vmCID)
	}

	if !node.PersistentDisk.IsAttached {
		container := rackhdapi.PersistentDiskSettingsContainer{
			PersistentDisk: node.PersistentDisk,
		}
		container.PersistentDisk.IsAttached = true

		bodyBytes, err := json.Marshal(container)
		if err != nil {
			return err
		}

		err = rackhdapi.PatchNode(c, node.ID, bodyBytes)
		if err != nil {
			return err
		}
	}

	return nil
}
