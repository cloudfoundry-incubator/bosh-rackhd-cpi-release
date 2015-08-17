package vm

import (
	bwcstem "github.com/cppforlife/baremetal_cpi/stemcell"
)

type Creator interface {
	// Create takes an agent id and creates a VM with provided configuration
	Create(string, bwcstem.Stemcell, Networks, Environment) (VM, error)
}

type Finder interface {
	Find(string) (VM, bool, error)
}

type VM interface {
	ID() string

	Delete() error
}

type Environment map[string]interface{}
