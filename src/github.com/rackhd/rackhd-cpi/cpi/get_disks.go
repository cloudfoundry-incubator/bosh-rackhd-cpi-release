package cpi

import (
	"errors"
	"reflect"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

func GetDisks(c config.Cpi, extInput bosh.MethodArguments) ([]string, error) {
	var vmCID string

	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(vmCID) {
		return nil, errors.New("Received unexpected type for vm cid")
	}

	vmCID = extInput[0].(string)

	node, err := rackhdapi.GetNodeByVMCID(c, vmCID)
	if err != nil {
		return nil, err
	}

	if node.PersistentDisk.DiskCID != "" {
		result := make([]string, 1)
		result[0] = node.PersistentDisk.DiskCID
		return result, nil
	} else {
		return make([]string, 0), nil
	}
}
