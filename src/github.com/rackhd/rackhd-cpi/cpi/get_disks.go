package cpi

import (
	"errors"
	"reflect"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

// GetDisks returns the persistent disk
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

	for _, tag := range node.Tags {
		if strings.HasPrefix(tag, DiskCIDTagPrefix) {
			return []string{tag}, nil
		}
	}

	return []string{}, nil
}
