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

	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if node.CPI.PersistentDisk.DiskCID == diskCID {
			if node.CPI.PersistentDisk.IsAttached {
				return fmt.Errorf("Disk: %s is attached\n", diskCID)
			}

			if node.CPI.VMCID != vmCID {
				return fmt.Errorf("Disk %s does not belong to VM %s\n", diskCID, vmCID)
			}

			node.CPI.PersistentDisk.IsAttached = true
			body := node.CPI
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return err
			}

			rackhdapi.PatchNode(c, node.ID, bodyBytes)
			return nil
		}
	}

	return fmt.Errorf("Disk: %s not found\n", diskCID)
}
