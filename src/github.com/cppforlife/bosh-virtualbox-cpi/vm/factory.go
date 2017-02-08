package vm

import (
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

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

	agentOptions AgentOptions

	logTag string
	logger boshlog.Logger
}

func NewFactory(
	opts FactoryOpts,
	uuidGen boshuuid.Generator,
	driver driver.Driver,
	runner driver.Runner,
	diskFactory bdisk.Factory,
	agentOptions AgentOptions,
	logger boshlog.Logger,
) Factory {
	return Factory{
		opts:    opts,
		uuidGen: uuidGen,

		driver:      driver,
		runner:      runner,
		diskFactory: diskFactory,

		agentOptions: agentOptions,

		logTag: "vm.Factory",
		logger: logger,
	}
}

func (f Factory) Create(agentID string, stemcell bstem.Stemcell, props VMProps, networks Networks, env Environment) (VM, error) {
	host := Host{bnet.NewNetworks(f.driver, f.logger)}

	if f.opts.AutoEnableNetworks {
		err := host.EnableNetworks(networks)
		if err != nil {
			return nil, bosherr.WrapError(err, "Enabling networks")
		}
	}

	vm, err := f.newClonedVM(stemcell)
	if err != nil {
		return nil, err
	}

	err = vm.SetProps(props)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, err
	}

	networksWithMACs, err := vm.ConfigureNICs(networks, host)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Configuring NICs")
	}

	initialAgentEnv := NewAgentEnvForVM(agentID, vm.ID(), networksWithMACs, env, f.agentOptions)

	err = vm.ConfigureAgent(initialAgentEnv)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Initial agent configuration")
	}

	ephemeralDisk, err := f.diskFactory.Create(props.EphemeralDisk)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Creating ephemeral disk")
	}

	err = vm.AttachEphemeralDisk(ephemeralDisk)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Attaching ephemeral disk")
	}

	err = vm.Start(props.GUI)
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

func (f Factory) newVM(id string) VMImpl {
	pdsOpts := bpds.PortDevicesOpts{Controller: f.opts.StorageController}
	portDevices := bpds.NewPortDevices(id, pdsOpts, f.driver, f.logger)
	store := NewStore(filepath.Join(f.opts.DirPath, id), f.runner)
	return NewVMImpl(id, portDevices, store, f.driver, f.logger)
}

func (f Factory) Find(id string) (VM, error) {
	return f.newVM(id), nil
}

func (f Factory) newClonedVM(stemcell bstem.Stemcell) (VMImpl, error) {
	cloneIDInternal, err := f.uuidGen.Generate()
	if err != nil {
		return VMImpl{}, bosherr.WrapError(err, "Generating clone VM id")
	}

	cloneID := "vm-" + cloneIDInternal

	_, err = f.driver.Execute(
		"clonevm", stemcell.ID(),
		"--snapshot", stemcell.SnapshotName(),
		"--options", "link",
		"--name", cloneID, // extra non-conflicting
		"--uuid", cloneIDInternal,
		"--register",
	)
	if err != nil {
		return VMImpl{}, bosherr.WrapError(err, "Cloning VM")
	}

	return f.newVM(cloneID), nil
}
