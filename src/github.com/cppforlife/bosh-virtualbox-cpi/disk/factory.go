package disk

import (
	"path/filepath"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type Factory struct {
	dirPath string
	uuidGen boshuuid.Generator

	driver driver.Driver
	runner driver.Runner

	logger boshlog.Logger
}

func NewFactory(
	dirPath string,
	uuidGen boshuuid.Generator,
	driver driver.Driver,
	runner driver.Runner,
	logger boshlog.Logger,
) Factory {
	return Factory{
		dirPath: dirPath,
		uuidGen: uuidGen,

		driver: driver,
		runner: runner,

		logger: logger,
	}
}

func (f Factory) Create(size int) (Disk, error) {
	id, err := f.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating disk id")
	}

	id = "disk-" + id

	disk := f.newDisk(apiv1.NewDiskCID(id))

	_, _, err = f.runner.Execute("mkdir", "-p", disk.Path())
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating disk parent")
	}

	_, err = f.driver.Execute(
		"createhd",
		"--filename", disk.VMDKPath(),
		"--size", strconv.Itoa(size),
		"--format", "VMDK",
		"--variant", "Standard",
	)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating disk")
	}

	return disk, nil
}

func (f Factory) Find(cid apiv1.DiskCID) (Disk, error) {
	return f.newDisk(cid), nil
}

func (f Factory) newDisk(cid apiv1.DiskCID) DiskImpl {
	diskPath := filepath.Join(f.dirPath, cid.AsString())
	return NewDiskImpl(cid, diskPath, f.runner, f.logger)
}
