package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type HasVM struct {
	vmFinder bvm.Finder
}

func NewHasVM(vmFinder bvm.Finder) HasVM {
	return HasVM{vmFinder: vmFinder}
}

func (a HasVM) Run(vmCID VMCID) (bool, error) {
	vm, err := a.vmFinder.Find(string(vmCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	return vm.Exists()
}
