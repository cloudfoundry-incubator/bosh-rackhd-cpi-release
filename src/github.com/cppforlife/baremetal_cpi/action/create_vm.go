package action

import (
	//bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"fmt"

	bwcstem "github.com/cppforlife/baremetal_cpi/stemcell"
	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
	httpclient "github.com/cloudfoundry/bosh-utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type CreateVM struct {
	stemcellFinder bwcstem.Finder
	vmCreator      bwcvm.Creator
}

type VMCloudProperties struct{}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bwcstem.Finder, vmCreator bwcvm.Creator) CreateVM {
	return CreateVM{
		stemcellFinder: stemcellFinder,
		vmCreator:      vmCreator,
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, _ VMCloudProperties, networks Networks, _ []DiskCID, env Environment) (VMCID, error) {
	//TODO implement full create vm using apis
	return VMCID("some-vm"), nil
}
