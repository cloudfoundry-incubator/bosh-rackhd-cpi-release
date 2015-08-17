package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
)

type HasVM struct {
	vmFinder bwcvm.Finder
}

func NewHasVM(vmFinder bwcvm.Finder) HasVM {
	return HasVM{vmFinder: vmFinder}
}

func (a HasVM) Run(vmCID VMCID) (bool, error) {
	_, found, err := a.vmFinder.Find(string(vmCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	return found, nil
}
