package disk

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type DiskImpl struct {
	cid  apiv1.DiskCID
	path string

	runner driver.Runner
	logger boshlog.Logger
}

func NewDiskImpl(
	cid apiv1.DiskCID,
	path string,
	runner driver.Runner,
	logger boshlog.Logger,
) DiskImpl {
	return DiskImpl{cid, path, runner, logger}
}

func (d DiskImpl) ID() apiv1.DiskCID { return d.cid }

func (d DiskImpl) Path() string { return d.path }

func (d DiskImpl) VMDKPath() string {
	return filepath.Join(d.path, "disk.vmdk")
}

func (d DiskImpl) Exists() (bool, error) {
	_, _, err := d.runner.Execute("ls", d.path)
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Checking disk '%s'", d.path)
	}

	// todo check status

	return true, nil
}

func (d DiskImpl) Delete() error {
	_, _, err := d.runner.Execute("rm", "-rf", d.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting disk '%s'", d.path)
	}

	return nil
}
