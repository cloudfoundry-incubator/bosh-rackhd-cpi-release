package bosh

/*
  We hope that some version of the BOSH Director will provide a well defined
  input interface, namely a valid JSON map, however, at this time we have an
   array with heterogeneous elements
*/

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

type MethodArguments []interface{}

type CpiRequest struct {
	Method    string          `json:"method"`
	Arguments MethodArguments `json:"arguments"`
}
