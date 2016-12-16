package cpi

import (
	"errors"
	"reflect"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

// HasDisk checks whether the persistent disk with disk cid exists
func HasDisk(c config.Cpi, extInput bosh.MethodArguments) (bool, error) {
	var diskCID string

	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(diskCID) {
		return false, errors.New("Received unexpected type for disk cid")
	}

	diskCID = extInput[0].(string)

	if diskCID == "" {
		return false, nil
	}

	node, err := rackhdapi.GetNodeByTag(c, diskCID)
	if err != nil {
		return false, err
	}

	for _, tag := range node.Tags {
		if tag == diskCID {
			return true, nil
		}
	}
	return false, nil
}
