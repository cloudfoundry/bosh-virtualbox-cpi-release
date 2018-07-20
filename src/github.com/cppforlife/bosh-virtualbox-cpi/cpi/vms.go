package cpi

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
	bvm "github.com/cppforlife/bosh-virtualbox-cpi/vm"
)

type VMs struct {
	stemcellFinder bstem.Finder
	creator        bvm.Creator
	finder         bvm.Finder
}

func NewVMs(stemcellFinder bstem.Finder, creator bvm.Creator, finder bvm.Finder) VMs {
	return VMs{stemcellFinder, creator, finder}
}

func (a VMs) CreateVM(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	diskCIDs []apiv1.DiskCID, env apiv1.VMEnv) (apiv1.VMCID, error) {

	cid, _, err := a.CreateVMV2(agentID, stemcellCID, cloudProps, networks, diskCIDs, env)
	return cid, err
}

func (a VMs) CreateVMV2(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	_ []apiv1.DiskCID, env apiv1.VMEnv) (apiv1.VMCID, apiv1.Networks, error) {

	stemcell, err := a.stemcellFinder.Find(stemcellCID)
	if err != nil {
		return apiv1.VMCID{}, networks, bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	vm, err := a.creator.Create(agentID, stemcell, cloudProps, networks, env)
	if err != nil {
		return apiv1.VMCID{}, networks, bosherr.WrapErrorf(err, "Creating VM with agent ID '%s'", agentID)
	}

	return vm.ID(), networks, nil
}

func (a VMs) DeleteVM(cid apiv1.VMCID) error {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding vm '%s'", cid)
	}

	err = vm.Delete()
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting vm '%s'", cid)
	}

	return nil
}

func (a VMs) SetVMMetadata(cid apiv1.VMCID, metadata apiv1.VMMeta) error {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.SetMetadata(metadata)
}

func (a VMs) HasVM(cid apiv1.VMCID) (bool, error) {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.Exists()
}

func (a VMs) RebootVM(cid apiv1.VMCID) error {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.Reboot()
}

func (a VMs) GetDisks(cid apiv1.VMCID) ([]apiv1.DiskCID, error) {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.DiskIDs()
}

func (a VMs) CalculateVMCloudProperties(res apiv1.VMResources) (apiv1.VMCloudProps, error) {
	return apiv1.NewVMCloudPropsFromMap(map[string]interface{}{
		"memory":         res.RAM,
		"cpus":           res.CPU,
		"ephemeral_disk": res.EphemeralDiskSize,
	}), nil
}
