package cpi

import (
	"errors"
	"fmt"
)

const (
	CREATE_VM          = "create_vm"
	DELETE_VM          = "delete_vm"
	HAS_VM             = "has_vm"
	REBOOT_VM          = "reboot_vm"
	SET_VM_METADATA    = "set_vm_metadata"
	CONFIGURE_NETWORKS = "configure_networks"

	CREATE_STEMCELL = "create_stemcell"
	DELETE_STEMCELL = "delete_stemcell"

	CREATE_DISK     = "create_disk"
	DELETE_DISK     = "delete_disk"
	ATTACH_DISK     = "attach_disk"
	DETACH_DISK     = "detach_disk"
	HAS_DISK        = "has_disk"
	GET_DISKS       = "get_disks"
	SNAPSHOT_DISK   = "snapshot_disk"
	DELETE_SNAPSHOT = "delete_snapshot"

	CURRENT_VM_ID = "current_vm_id"
)

var cpiMethods = map[string]bool{
	CREATE_VM:          true,
	DELETE_VM:          true,
	HAS_VM:             true,
	REBOOT_VM:          false,
	SET_VM_METADATA:    true,
	CONFIGURE_NETWORKS: false,
	CREATE_STEMCELL:    true,
	DELETE_STEMCELL:    true,
	CREATE_DISK:        true,
	DELETE_DISK:        true,
	ATTACH_DISK:        true,
	DETACH_DISK:        true,
	HAS_DISK:           true,
	GET_DISKS:          true,
	SNAPSHOT_DISK:      false,
	DELETE_SNAPSHOT:    false,
	CURRENT_VM_ID:      false,
}

func ImplementsMethod(method string) (bool, error) {
	implemented, exists := cpiMethods[method]
	if !exists {
		// WARNING: implemented is undefined in this case
		return implemented, errors.New(fmt.Sprintf("Invalid Method: %s", method))
	}

	return implemented, nil
}
