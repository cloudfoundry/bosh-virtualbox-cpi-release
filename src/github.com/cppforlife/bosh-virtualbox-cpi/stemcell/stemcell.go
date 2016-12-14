package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type StemcellImpl struct {
	id   string
	path string

	driver driver.Driver
	runner driver.Runner

	logger boshlog.Logger
}

func NewStemcellImpl(
	id string,
	path string,
	driver driver.Driver,
	runner driver.Runner,
	logger boshlog.Logger,
) StemcellImpl {
	return StemcellImpl{id: id, path: path, driver: driver, runner: runner, logger: logger}
}

func (s StemcellImpl) ID() string { return s.id }

func (s StemcellImpl) SnapshotName() string { return "prepared-clone" }

func (s StemcellImpl) Prepare() error {
	_, err := s.driver.Execute("snapshot", s.id, "take", s.SnapshotName())
	if err != nil {
		return bosherr.WrapErrorf(err, "Preparing for future cloning")
	}

	return nil
}

func (s StemcellImpl) Exists() (bool, error) {
	output, err := s.driver.Execute("showvminfo", s.id, "--machinereadable")
	if err != nil {
		if s.driver.IsMissingVMErr(output) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s StemcellImpl) Delete() error {
	output, err := s.driver.Execute("unregistervm", s.id, "--delete")
	if err != nil {
		if !s.driver.IsMissingVMErr(output) {
			return bosherr.WrapErrorf(err, "Unregistering stemcell VM")
		}
	}

	_, _, err = s.runner.Execute("rm", "-rf", s.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting stemcell '%s'", s.path)
	}

	return nil
}
