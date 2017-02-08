package vm

import (
	"encoding/json"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
	bpds "github.com/cppforlife/bosh-virtualbox-cpi/vm/portdevices"
)

type VMImpl struct {
	id          string
	portDevices bpds.PortDevices
	store       Store

	driver driver.Driver
	logger boshlog.Logger
}

func NewVMImpl(
	id string,
	portDevices bpds.PortDevices,
	store Store,
	driver driver.Driver,
	logger boshlog.Logger,
) VMImpl {
	return VMImpl{
		id:          id,
		portDevices: portDevices,
		store:       store,
		driver:      driver,
		logger:      logger,
	}
}

func (vm VMImpl) ID() string { return vm.id }

func (vm VMImpl) SetProps(props VMProps) error {
	_, err := vm.driver.Execute(
		"modifyvm", vm.id,
		"--name", vm.id,
		"--memory", strconv.Itoa(props.Memory),
		"--cpus", strconv.Itoa(props.CPUs),
		// Using minimal paravirtualization provider to avoid CPU lockups
		"--paravirtprovider", props.ParavirtProvider,
	)
	return err
}

func (vm VMImpl) SetMetadata(meta VMMetadata) error {
	bytes, err := json.Marshal(meta)
	if err != nil {
		return bosherr.WrapError(err, "Marshaling VM metadata")
	}

	err = vm.store.Put("metadata.json", bytes)
	if err != nil {
		return bosherr.WrapError(err, "Saving VM metadata")
	}

	return nil
}

func (vm VMImpl) ConfigureNICs(nets Networks, host Host) (Networks, error) {
	return NICs{vm.driver, vm.ID()}.Configure(nets, host)
}

func (vm VMImpl) Delete() error {
	err := vm.HaltIfRunning()
	if err != nil {
		return err
	}

	// todo is this necessary?
	err = vm.detachPersistentDisks()
	if err != nil {
		return err
	}

	_, err = vm.driver.Execute("unregistervm", vm.id, "--delete")
	if err != nil {
		return err
	}

	return vm.store.Delete()
}
