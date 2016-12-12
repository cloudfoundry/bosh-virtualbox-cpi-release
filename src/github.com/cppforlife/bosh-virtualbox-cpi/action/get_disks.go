package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type GetDisks struct {
	vmFinder bvm.Finder
}

func NewGetDisks(vmFinder bvm.Finder) GetDisks {
	return GetDisks{vmFinder: vmFinder}
}

func (a GetDisks) Run(vmCID VMCID) ([]string, error) {
	vm, err := a.vmFinder.Find(string(vmCID))
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	return vm.DiskIDs()
}
