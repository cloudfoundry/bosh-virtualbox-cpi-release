package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type StemcellImpl struct {
	cid  apiv1.StemcellCID
	path string

	driver driver.Driver
	runner driver.Runner

	logger boshlog.Logger
}

func NewStemcellImpl(
	cid apiv1.StemcellCID,
	path string,
	driver driver.Driver,
	runner driver.Runner,
	logger boshlog.Logger,
) StemcellImpl {
	return StemcellImpl{cid, path, driver, runner, logger}
}

func (s StemcellImpl) ID() apiv1.StemcellCID { return s.cid }

func (s StemcellImpl) SnapshotName() string { return "prepared-clone" }

func (s StemcellImpl) Prepare() error {
	_, err := s.driver.Execute("snapshot", s.cid.AsString(), "take", s.SnapshotName())
	if err != nil {
		return bosherr.WrapErrorf(err, "Preparing for future cloning")
	}

	return nil
}

func (s StemcellImpl) Exists() (bool, error) {
	output, err := s.driver.Execute("showvminfo", s.cid.AsString(), "--machinereadable")
	if err != nil {
		if s.driver.IsMissingVMErr(output) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s StemcellImpl) Delete() error {
	output, err := s.driver.Execute("unregistervm", s.cid.AsString(), "--delete")
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
