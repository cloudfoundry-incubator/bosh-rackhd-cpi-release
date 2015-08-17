package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bwcstem "github.com/cppforlife/baremetal_cpi/stemcell"
)

type CreateStemcell struct {
	stemcellImporter bwcstem.Importer
}

type CreateStemcellCloudProps struct{}

func NewCreateStemcell(stemcellImporter bwcstem.Importer) CreateStemcell {
	return CreateStemcell{stemcellImporter: stemcellImporter}
}

func (a CreateStemcell) Run(imagePath string, _ CreateStemcellCloudProps) (StemcellCID, error) {
	stemcell, err := a.stemcellImporter.ImportFromPath(imagePath)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Importing stemcell from '%s'", imagePath)
	}

	return StemcellCID(stemcell.ID()), nil
}
