package cpi

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

// DetachDisk detaches disk from vm
func DetachDisk(c config.Cpi, extInput bosh.MethodArguments) error {
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
		return err
	}

	for _, tag := range node.Tags {
		if strings.HasPrefix(tag, DiskCIDTagPrefix) {
			if tag != diskCID {
				return fmt.Errorf("another disk is attached to VM %s", vmCID)
			}

			if !node.PersistentDisk.IsAttached {
				return fmt.Errorf("disk: %s is already detached to VM %s", diskCID, vmCID)
			}

			return rackhdapi.MakeDiskRequest(c, node, false)
		}
	}

	return fmt.Errorf("disk: %s was not found on VM %s", diskCID, vmCID)
}
