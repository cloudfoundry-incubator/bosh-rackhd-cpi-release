package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
)

type ConcreteFactoryOptions struct {
	User       string
	Port       int
	PrivateKey string

	Machines map[string]string

	StemcellsDir string

	Agent bwcvm.AgentOptions
}

func (o ConcreteFactoryOptions) Validate() error {
	if o.User == "" {
		return bosherr.Error("Must provide non-empty User")
	}

	if o.Port == 0 {
		return bosherr.Error("Must provide non-zero Port")
	}

	if o.PrivateKey == "" {
		return bosherr.Error("Must provide non-empty PrivateKey")
	}

	if len(o.Machines) == 0 {
		return bosherr.Error("Must provide at least one machine in Machines")
	}

	if o.StemcellsDir == "" {
		return bosherr.Error("Must provide non-empty StemcellsDir")
	}

	err := o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	return nil
}
