package cpi

import (
	"errors"
	"fmt"
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

	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if node.CPI.VMCID == vmCID {
			var result []string = []string{}
			if node.CPI.PersistentDisk.DiskCID != "" {
				result = []string{node.CPI.PersistentDisk.DiskCID}
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("VM: %s not found\n", vmCID)
}
