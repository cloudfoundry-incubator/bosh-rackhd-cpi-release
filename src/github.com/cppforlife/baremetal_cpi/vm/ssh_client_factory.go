package vm

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	gossh "golang.org/x/crypto/ssh"
)

type SSHClientFactory struct {
	uuidGen boshuuid.Generator

	user       string
	port       int
	privateKey string

	machineIDToIP map[string]string

	logger boshlog.Logger
}

func NewSSHClientFactory(
	uuidGen boshuuid.Generator,
	user string,
	port int,
	privateKey string,
	machineIDToIP map[string]string,
	logger boshlog.Logger,
) SSHClientFactory {
	return SSHClientFactory{
		uuidGen: uuidGen,

		user:       user,
		port:       port,
		privateKey: privateKey,

		machineIDToIP: machineIDToIP,

		logger: logger,
	}
}

func (f SSHClientFactory) New(machineID string) (SSHClient, error) {
	ip, found := f.machineIDToIP[machineID]
	if !found {
		return SSHClient{}, bosherr.Errorf("Machine ID '%s' is not found", machineID)
	}

	// todo move onto config?
	key, err := gossh.ParsePrivateKey([]byte(f.privateKey))
	if err != nil {
		return SSHClient{}, bosherr.WrapError(err, "Parsing private key")
	}

	config := &gossh.ClientConfig{
		User: f.user,
		Auth: []gossh.AuthMethod{gossh.PublicKeys(key)}, // todo password auth?
	}

	client, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", ip, f.port), config)
	if err != nil {
		return SSHClient{}, bosherr.WrapErrorf(err,
			"Connecting to IP '%s:%d' for machine ID '%s'", ip, f.port, machineID)
	}

	return NewSSHClient(client, f.uuidGen, f.logger), nil
}
