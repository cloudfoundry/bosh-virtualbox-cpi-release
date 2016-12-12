package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
)

type DeleteDisk struct {
	diskFinder bdisk.Finder
}

func NewDeleteDisk(diskFinder bdisk.Finder) DeleteDisk {
	return DeleteDisk{diskFinder: diskFinder}
}

func (a DeleteDisk) Run(diskCID DiskCID) (interface{}, error) {
	disk, err := a.diskFinder.Find(string(diskCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	err = disk.Delete()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Deleting disk '%s'", diskCID)
	}

	return nil, nil
}
