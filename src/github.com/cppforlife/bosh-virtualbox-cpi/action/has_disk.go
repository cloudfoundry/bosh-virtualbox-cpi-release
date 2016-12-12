package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
)

type HasDisk struct {
	diskFinder bdisk.Finder
}

func NewHasDisk(diskFinder bdisk.Finder) HasDisk {
	return HasDisk{diskFinder: diskFinder}
}

func (a HasDisk) Run(diskCID DiskCID) (bool, error) {
	disk, err := a.diskFinder.Find(string(diskCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	return disk.Exists()
}
