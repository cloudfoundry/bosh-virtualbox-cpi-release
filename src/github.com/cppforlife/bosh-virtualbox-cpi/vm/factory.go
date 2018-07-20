package vm

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bdisk "github.com/cppforlife/bosh-virtualbox-cpi/disk"
	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
	bstem "github.com/cppforlife/bosh-virtualbox-cpi/stemcell"
	bnet "github.com/cppforlife/bosh-virtualbox-cpi/vm/network"
	bpds "github.com/cppforlife/bosh-virtualbox-cpi/vm/portdevices"
)

type FactoryOpts struct {
	DirPath            string
	StorageController  string
	AutoEnableNetworks bool
}

type Factory struct {
	opts    FactoryOpts
	uuidGen boshuuid.Generator

	driver      driver.Driver
	runner      driver.Runner
	diskFactory bdisk.Factory

	agentOptions       apiv1.AgentOptions
	stemcellAPIVersion apiv1.StemcellAPIVersion

	logTag string
	logger boshlog.Logger
}

func NewFactory(
	opts FactoryOpts,
	uuidGen boshuuid.Generator,
	driver driver.Driver,
	runner driver.Runner,
	diskFactory bdisk.Factory,
	agentOptions apiv1.AgentOptions,
	stemcellAPIVersion apiv1.StemcellAPIVersion,
	logger boshlog.Logger,
) Factory {
	return Factory{
		opts:    opts,
		uuidGen: uuidGen,

		driver:      driver,
		runner:      runner,
		diskFactory: diskFactory,

		agentOptions:       agentOptions,
		stemcellAPIVersion: stemcellAPIVersion,

		logTag: "vm.Factory",
		logger: logger,
	}
}

func (f Factory) Create(
	agentID apiv1.AgentID,
	stemcell bstem.Stemcell,
	props apiv1.VMCloudProps,
	networks apiv1.Networks,
	env apiv1.VMEnv,
) (VM, error) {

	host := Host{bnet.NewNetworks(f.driver, f.logger)}

	vmProps, err := NewVMProps(props)
	if err != nil {
		return nil, err
	}

	vmNetworks, err := NewNetworks(networks)
	if err != nil {
		return nil, err
	}

	if f.opts.AutoEnableNetworks {
		err := host.EnableNetworks(vmNetworks)
		if err != nil {
			return nil, bosherr.WrapError(err, "Enabling networks")
		}
	}

	vm, err := f.newClonedVM(stemcell)
	if err != nil {
		return nil, err
	}

	err = vm.SetProps(vmProps)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, err
	}

	err = vm.ConfigureNICs(vmNetworks, host)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Configuring NICs")
	}

	initialAgentEnv := apiv1.NewAgentEnvFactory().ForVM(
		agentID, vm.ID(), vmNetworks.AsNetworks(), env, f.agentOptions)

	initialAgentEnv.AttachSystemDisk(apiv1.NewDiskHintFromString("0"))

	err = vm.ConfigureAgent(initialAgentEnv)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Initial agent configuration")
	}

	ephemeralDisk, err := f.diskFactory.Create(vmProps.EphemeralDisk)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Creating ephemeral disk")
	}

	err = vm.AttachEphemeralDisk(ephemeralDisk)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Attaching ephemeral disk")
	}

	err = vm.Start(vmProps.GUI)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Starting VM")
	}

	return vm, nil
}

func (f Factory) cleanUpPartialCreate(vm VM) {
	err := vm.Delete()
	if err != nil {
		f.logger.Error(f.logTag, "Failed to clean up partially created VM: %s", err)
	}
}

func (f Factory) newVM(cid apiv1.VMCID) VMImpl {
	pdsOpts := bpds.PortDevicesOpts{Controller: f.opts.StorageController}
	portDevices := bpds.NewPortDevices(cid, pdsOpts, f.driver, f.logger)
	store := NewStore(filepath.Join(f.opts.DirPath, cid.AsString()), f.runner)
	return NewVMImpl(cid, portDevices, store, f.stemcellAPIVersion, f.driver, f.logger)
}

func (f Factory) Find(cid apiv1.VMCID) (VM, error) {
	return f.newVM(cid), nil
}

func (f Factory) newClonedVM(stemcell bstem.Stemcell) (VMImpl, error) {
	cloneIDInternal, err := f.uuidGen.Generate()
	if err != nil {
		return VMImpl{}, bosherr.WrapError(err, "Generating clone VM id")
	}

	cloneID := "vm-" + cloneIDInternal

	_, err = f.driver.Execute(
		"clonevm", stemcell.ID().AsString(),
		"--snapshot", stemcell.SnapshotName(),
		"--options", "link",
		"--name", cloneID, // extra non-conflicting
		"--uuid", cloneIDInternal,
		"--register",
	)
	if err != nil {
		return VMImpl{}, bosherr.WrapError(err, "Cloning VM")
	}

	return f.newVM(apiv1.NewVMCID(cloneID)), nil
}
