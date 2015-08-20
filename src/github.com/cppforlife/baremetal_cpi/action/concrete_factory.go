package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bwcstem "github.com/cppforlife/baremetal_cpi/stemcell"
	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
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

	stemcellFinder := bwcstem.NewFSFinder(options.StemcellsDir, fs, logger)

	sshClientFactory := bwcvm.NewSSHClientFactory(
		uuidGen,
		options.User,
		options.Port,
		options.PrivateKey,
		options.Machines,
		logger,
	)

	vmCreator := bwcvm.NewSSHCreator(
		uuidGen,
		sshClientFactory,
		options.Agent,
		logger,
	)

	vmFinder := bwcvm.NewSSHFinder(sshClientFactory, logger)

	return concreteFactory{
		availableActions: map[string]Action{
			// Stemcell management
			"create_stemcell": NewCreateStemcell(options.APIServer, logger),
			"delete_stemcell": NewDeleteStemcell(stemcellFinder),

			// VM management
			"create_vm":          NewCreateVM(stemcellFinder, vmCreator),
			"delete_vm":          NewDeleteVM(vmFinder),
			"has_vm":             NewHasVM(vmFinder),
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
		return nil, bosherr.Errorf("Could not create action with method %s", method)
	}

	return action, nil
}
