package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type RebootVM struct {
	vmFinder bvm.Finder
}

func NewRebootVM(vmFinder bvm.Finder) RebootVM {
	return RebootVM{vmFinder: vmFinder}
}

func (a RebootVM) Run(vmCID VMCID) (interface{}, error) {
	vm, err := a.vmFinder.Find(string(vmCID))
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	return "", vm.Reboot()
}
