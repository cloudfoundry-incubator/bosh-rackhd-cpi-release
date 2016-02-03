package cpi

import (
	"errors"
	"fmt"

	"github.com/rackhd/rackhd-cpi/bosh"
)

var cpiMethods = map[string]bool{
	bosh.CREATE_VM:          true,
	bosh.DELETE_VM:          true,
	bosh.HAS_VM:             true,
	bosh.REBOOT_VM:          false,
	bosh.SET_VM_METADATA:    true,
	bosh.CONFIGURE_NETWORKS: false,
	bosh.CREATE_STEMCELL:    true,
	bosh.DELETE_STEMCELL:    true,
	bosh.CREATE_DISK:        true,
	bosh.DELETE_DISK:        true,
	bosh.ATTACH_DISK:        true,
	bosh.DETACH_DISK:        true,
	bosh.HAS_DISK:           true,
	bosh.GET_DISKS:          true,
	bosh.SNAPSHOT_DISK:      false,
	bosh.DELETE_SNAPSHOT:    false,
	bosh.CURRENT_VM_ID:      false,
}

func ImplementsMethod(method string) (bool, error) {
	implemented, exists := cpiMethods[method]
	if !exists {
		// WARNING: implemented is undefined in this case
		return implemented, errors.New(fmt.Sprintf("Invalid Method: %s", method))
	}

	return implemented, nil
}
