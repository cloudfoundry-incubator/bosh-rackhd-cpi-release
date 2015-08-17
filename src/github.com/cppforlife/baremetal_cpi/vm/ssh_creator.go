package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	bwcstem "github.com/cppforlife/baremetal_cpi/stemcell"
)

type SSHCreator struct {
	uuidGen          boshuuid.Generator
	sshClientFactory SSHClientFactory

	agentOptions AgentOptions

	logTag string
	logger boshlog.Logger
}

func NewSSHCreator(
	uuidGen boshuuid.Generator,
	sshClientFactory SSHClientFactory,
	agentOptions AgentOptions,
	logger boshlog.Logger,
) SSHCreator {
	return SSHCreator{
		uuidGen:          uuidGen, // todo remove
		sshClientFactory: sshClientFactory,

		agentOptions: agentOptions,

		logTag: "SSHCreator",
		logger: logger,
	}
}

func (c SSHCreator) Create(agentID string, stemcell bwcstem.Stemcell, networks Networks, env Environment) (VM, error) {
	machineID := c.selectMachine()

	sshClient, err := c.sshClientFactory.New(machineID)
	if err != nil {
		return SSHVM{}, bosherr.WrapErrorf(err,
			"Establishing connection to machine ID '%s'", machineID)
	}

	agentEnvService := NewFSAgentEnvService(sshClient, c.logger)

	vm := NewSSHVM(machineID, sshClient, agentEnvService, c.logger, true)

	err = vm.SetUpStemcell(stemcell.Path())
	if err != nil {
		c.deleteVM(vm)
		return SSHVM{}, bosherr.WrapError(err, "Setting up stemcell")
	}

	agentEnv := NewAgentEnvForVM(agentID, machineID, networks, env, c.agentOptions)

	// todo should this be inside vm?
	err = agentEnvService.Update(agentEnv)
	if err != nil {
		c.deleteVM(vm)
		return SSHVM{}, bosherr.WrapError(err, "Updating vm's agent env")
	}

	err = vm.StartAgent()
	if err != nil {
		c.deleteVM(vm)
		return SSHVM{}, err
	}

	return vm, nil
}

func (c SSHCreator) deleteVM(vm VM) {
	err := vm.Delete()
	if err != nil {
		c.logger.Error(c.logTag, "Failed to delete machine ID '%s': %s", vm.ID(), err)
	}
}

func (c SSHCreator) selectMachine() string {
	machineID := "some-vm" // todo better selection, based on IPs, disk IDs?

	return machineID
}
