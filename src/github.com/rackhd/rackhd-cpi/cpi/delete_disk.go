package cpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

// DeleteDisk deprovisions disk
func DeleteDisk(c config.Cpi, extInput bosh.MethodArguments) error {
	var diskCID string

	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(diskCID) {
		return errors.New("Received unexpected type for disk cid")
	}
	diskCID = extInput[0].(string)

	node, err := rackhdapi.GetNodeByTag(c, diskCID)

	if err != nil {
		return err
	}
	if node.PersistentDisk.IsAttached {
		return fmt.Errorf("disk: %s is attached", diskCID)
	}

	container := models.PersistentDiskSettingsContainer{
		PersistentDisk: models.PersistentDiskSettings{},
	}
	bodyBytes, err := json.Marshal(container)
	if err != nil {
		return err
	}

	err = rackhdapi.PatchNode(c, node.ID, bodyBytes)
	if err != nil {
		return fmt.Errorf("error deleting disk metadata %s: %s", diskCID, err)
	}

	err = rackhdapi.DeleteTag(c, node.ID, diskCID)
	if err != nil {
		return fmt.Errorf("error deleting disk cid tag %s: %s", diskCID, err)
	}

	for _, tag := range node.Tags {
		if strings.HasPrefix(tag, VMCIDTagPrefix) {
			return nil
		}
	}

	err = rackhdapi.ReleaseNode(c, node.ID)
	if err != nil {
		return fmt.Errorf("error releasing node after delete disk %s: %v", diskCID, err)
	}
	return nil
}
