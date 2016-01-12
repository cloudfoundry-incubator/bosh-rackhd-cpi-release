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

func DeleteDisk(c config.Cpi, extInput bosh.MethodArguments) error {
	var diskCID string

	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(diskCID) {
		return errors.New("Received unexpected type for disk cid")
	}

	diskCID = extInput[0].(string)

	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if node.PersistentDisk.DiskCID == diskCID {
			if node.PersistentDisk.IsAttached {
				return fmt.Errorf("Disk: %s is attached\n", diskCID)
			}

			container := rackhdapi.PersistentDiskSettingsContainer{
				PersistentDisk: rackhdapi.PersistentDiskSettings{},
			}
			bodyBytes, err := json.Marshal(container)
			if err != nil {
				return err
			}

			rackhdapi.PatchNode(c, node.ID, bodyBytes)
			return nil
		}
	}

	return fmt.Errorf("Disk: %s not found\n", diskCID)
}
