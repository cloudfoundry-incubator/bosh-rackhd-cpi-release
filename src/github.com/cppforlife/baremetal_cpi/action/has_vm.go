package action

import (
	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
)

type HasVM struct {
	vmFinder bwcvm.Finder
}

func NewHasVM(vmFinder bwcvm.Finder) HasVM {
	return HasVM{vmFinder: vmFinder}
}

func (a HasVM) Run(vmCID VMCID) (bool, error) {
	//TODO implement
	return false, nil
}
