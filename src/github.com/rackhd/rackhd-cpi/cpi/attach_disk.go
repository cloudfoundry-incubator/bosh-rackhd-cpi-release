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

// AttachDisk attack a disk to a machine
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
		return fmt.Errorf("VM: %s not found", vmCID)
	}

	var attachedDiskCID string
	for _, tag := range node.Tags {
		if strings.HasPrefix(tag, DiskCIDTagPrefix) {
			if tag == diskCID && node.PersistentDisk.IsAttached {
				return nil
			}
			attachedDiskCID = tag
		}
	}

	if attachedDiskCID == "" {
		return fmt.Errorf("disk: %s not found on VM: %s", diskCID, vmCID)
	}

	if attachedDiskCID != diskCID {
		if node.PersistentDisk.IsAttached {
			return fmt.Errorf("node %s has persistent disk %s attached. Cannot attach additional disk %s", vmCID, attachedDiskCID, diskCID)
		}
		return fmt.Errorf("node %s has persistent disk %s, but detached. Cannot attach disk %s", vmCID, attachedDiskCID, diskCID)
	}

	err = rackhdapi.MakeDiskRequest(c, node, true)
	return err
}
