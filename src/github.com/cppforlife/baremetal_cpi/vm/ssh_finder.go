package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type SSHFinder struct {
	sshClientFactory SSHClientFactory

	logTag string
	logger boshlog.Logger
}

func NewSSHFinder(
	sshClientFactory SSHClientFactory,
	logger boshlog.Logger,
) SSHFinder {
	return SSHFinder{
		sshClientFactory: sshClientFactory,

		logTag: "SSHFinder",
		logger: logger,
	}
}

func (f SSHFinder) Find(id string) (VM, bool, error) {
	f.logger.Debug(f.logTag, "Finding machine with ID '%s'", id)

	sshClient, err := f.sshClientFactory.New(id)
	if err != nil {
		// todo proper boolean?
		return SSHVM{}, false, bosherr.WrapErrorf(err,
			"Establishing connection to machine ID '%s'", id)
	}

	agentEnvService := NewFSAgentEnvService(sshClient, f.logger)

	vm := NewSSHVM(id, sshClient, agentEnvService, f.logger, true)

	return vm, true, nil
}
