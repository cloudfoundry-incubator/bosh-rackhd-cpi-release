package stemcell

import (
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

const fsImporterLogTag = "FSImporter"

type FSImporter struct {
	dirPath string

	fs      boshsys.FileSystem
	uuidGen boshuuid.Generator

	logger boshlog.Logger
}

func NewFSImporter(
	dirPath string,
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) FSImporter {
	return FSImporter{
		dirPath: dirPath,

		fs:      fs,
		uuidGen: uuidGen,

		logger: logger,
	}
}

func (i FSImporter) ImportFromPath(imagePath string) (Stemcell, error) {
	i.logger.Debug(fsImporterLogTag, "Importing stemcell from path '%s'", imagePath)

	id, err := i.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating stemcell id")
	}

	err = i.fs.MkdirAll(i.dirPath, os.FileMode(0755))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Creating stemcell directory '%s'", i.dirPath)
	}

	stemcellPath := filepath.Join(i.dirPath, id)

	err = i.fs.CopyFile(imagePath, stemcellPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Copying stemcell '%s'", stemcellPath)
	}

	i.logger.Debug(fsImporterLogTag, "Imported stemcell from path '%s'", imagePath)

	return NewFSStemcell(id, stemcellPath, i.fs, i.logger), nil
}
