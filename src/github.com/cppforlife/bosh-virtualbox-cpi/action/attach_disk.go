package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type AttachDisk struct {
	vmFinder   bvm.Finder
	diskFinder bdisk.Finder
}

func NewAttachDisk(vmFinder bvm.Finder, diskFinder bdisk.Finder) AttachDisk {
	return AttachDisk{
		vmFinder:   vmFinder,
		diskFinder: diskFinder,
	}
}

func (a AttachDisk) Run(vmCID VMCID, diskCID DiskCID) (interface{}, error) {
	vm, err := a.vmFinder.Find(string(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	disk, err := a.diskFinder.Find(string(diskCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding disk '%s'", diskCID)
	}

	err = vm.AttachDisk(disk)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Attaching disk '%s' to VM '%s'", diskCID, vmCID)
	}

	return nil, nil
}
