package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"fmt"
)

type concreteFactory struct {
	availableActions map[string]Action
}

func NewConcreteFactory(
	fs boshsys.FileSystem,
	cmdRunner boshsys.CmdRunner,
	uuidGen boshuuid.Generator,
	options ConcreteFactoryOptions,
	logger boshlog.Logger,
) concreteFactory {
	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(options.APIServer, logger),
			"delete_stemcell": NewDeleteStemcell(options.APIServer, logger),

			// VM management
			"create_vm":          NewCreateVM(options.APIServer, options.Agent, logger),
			"delete_vm":          NewDeleteVM(options.APIServer, logger),
			"has_vm":             NewHasVM(options.APIServer, logger),
			"reboot_vm":          NewRebootVM(),
			"set_vm_metadata":    NewSetVMMetadata(),
			"configure_networks": NewConfigureNetworks(),

			// Disk management
			// "create_disk": NewCreateDisk(diskCreator),
			// "delete_disk": NewDeleteDisk(diskFinder),
			// "attach_disk": NewAttachDisk(vmFinder, diskFinder),
			// "detach_disk": NewDetachDisk(vmFinder, diskFinder),

			// Not implemented:
			//   current_vm_id
			//   snapshot_disk
			//   delete_snapshot
			//   get_disks
			//   ping
		},
	}
}

func (f concreteFactory) Create(method string) (Action, error) {
	action, found := f.availableActions[method]
	if !found {
		return nil, fmt.Errorf("Could not create action with method %s", method)
	}

	return action, nil
}
