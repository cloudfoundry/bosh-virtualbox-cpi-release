package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type CreateVM struct {
	stemcellFinder bstem.Finder
	vmCreator      bvm.Creator
}

type VMCloudProperties struct {
	Memory        int
	CPUs          int
	EphemeralDisk int `json:"ephemeral_disk"`

	GUI              bool
	ParavirtProvider string `json:"paravirtprovider"`
}

func (cp VMCloudProperties) AsVMProps() bvm.VMProps {
	props := bvm.VMProps{
		Memory:        512,
		CPUs:          1,
		EphemeralDisk: 5000,

		GUI:              cp.GUI,
		ParavirtProvider: "minimal", // KVM caused CPU lockups with 4+ kernel
	}
	if cp.Memory > 0 {
		props.Memory = cp.Memory
	}
	if cp.CPUs > 0 {
		props.CPUs = cp.CPUs
	}
	if cp.EphemeralDisk > 0 {
		props.EphemeralDisk = cp.EphemeralDisk
	}
	if len(cp.ParavirtProvider) > 0 {
		props.ParavirtProvider = cp.ParavirtProvider
	}
	return props
}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bstem.Finder, vmCreator bvm.Creator) CreateVM {
	return CreateVM{
		stemcellFinder: stemcellFinder,
		vmCreator:      vmCreator,
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, cloudProps VMCloudProperties, networks Networks, _ []DiskCID, env Environment) (VMCID, error) {
	stemcell, err := a.stemcellFinder.Find(string(stemcellCID))
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	vmNetworks := networks.AsVMNetworks()
	vmProps := cloudProps.AsVMProps()
	vmEnv := bvm.Environment(env)

	vm, err := a.vmCreator.Create(agentID, stemcell, vmProps, vmNetworks, vmEnv)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating VM with agent ID '%s'", agentID)
	}

	return VMCID(vm.ID()), nil
}
