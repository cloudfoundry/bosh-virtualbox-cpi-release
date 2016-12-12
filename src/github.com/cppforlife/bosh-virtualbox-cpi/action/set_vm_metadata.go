package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type SetVMMetadata struct {
	vm bvm.Finder
}

type VMMetadata map[string]interface{}

func NewSetVMMetadata(vm bvm.Finder) SetVMMetadata {
	return SetVMMetadata{vm: vm}
}

func (a SetVMMetadata) Run(vmCID VMCID, metadata VMMetadata) (interface{}, error) {
	vm, err := a.vm.Find(string(vmCID))
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM '%s'", vmCID)
	}

	return true, vm.SetMetadata(bvm.VMMetadata(metadata))
}
