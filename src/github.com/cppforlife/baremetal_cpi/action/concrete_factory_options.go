package action

import (
	bwcvm "github.com/cppforlife/baremetal_cpi/vm"

	"errors"
)

type ConcreteFactoryOptions struct {
	APIServer string
	Agent bwcvm.AgentOptions
}

func (o ConcreteFactoryOptions) Validate() error {
	if o.APIServer == "" {
		return errors.New("Must provide API Server IP")
	}

	err := o.Agent.Validate()
	if err != nil {
		return err
	}

	return nil
}
