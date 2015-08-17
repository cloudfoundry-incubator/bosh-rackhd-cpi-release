package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

const fsStemcellLogTag = "FSStemcell"

type FSStemcell struct {
	id   string
	path string

	fs     boshsys.FileSystem
	logger boshlog.Logger
}

func NewFSStemcell(id string, path string, fs boshsys.FileSystem, logger boshlog.Logger) FSStemcell {
	return FSStemcell{id: id, path: path, fs: fs, logger: logger}
}

func (s FSStemcell) ID() string { return s.id }

func (s FSStemcell) Path() string { return s.path }

func (s FSStemcell) Delete() error {
	s.logger.Debug(fsStemcellLogTag, "Deleting stemcell '%s'", s.id)

	err := s.fs.RemoveAll(s.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting stemcell directory '%s'", s.path)
	}

	return nil
}
