package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type DetachDisk struct {
	vmFinder   bvm.Finder
	diskFinder bdisk.Finder
}

func NewDetachDisk(vmFinder bvm.Finder, diskFinder bdisk.Finder) DetachDisk {
	return DetachDisk{
		vmFinder:   vmFinder,
		diskFinder: diskFinder,
	}
}

func (a DetachDisk) Run(vmCID VMCID, diskCID DiskCID) (interface{}, error) {
	vm, err := a.vmFinder.Find(string(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	disk, err := a.diskFinder.Find(string(diskCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	err = vm.DetachDisk(disk)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Detaching disk '%s' to VM '%s'", diskCID, vmCID)
	}

	return nil, nil
}
