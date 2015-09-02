package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
)

type ConcreteFactoryOptions struct {
	APIServer string
	Agent bwcvm.AgentOptions
}

func (o ConcreteFactoryOptions) Validate() error {
	if o.APIServer == "" {
		return bosherr.Error("Must provide API Server IP")
	}

	err := o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	return nil
}
