package vm

import (
	"encoding/json"
	"fmt"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
	bpds "github.com/cppforlife/bosh-virtualbox-cpi/vm/portdevices"
)

type VMImpl struct {
	cid         apiv1.VMCID
	portDevices bpds.PortDevices
	store       Store

	stemcellAPIVersion apiv1.StemcellAPIVersion

	driver driver.Driver
	logger boshlog.Logger
}

func NewVMImpl(
	cid apiv1.VMCID,
	portDevices bpds.PortDevices,
	store Store,
	stemcellAPIVersion apiv1.StemcellAPIVersion,
	driver driver.Driver,
	logger boshlog.Logger,
) VMImpl {
	return VMImpl{
		cid:                cid,
		portDevices:        portDevices,
		store:              store,
		stemcellAPIVersion: stemcellAPIVersion,
		driver:             driver,
		logger:             logger,
	}
}

func (vm VMImpl) ID() apiv1.VMCID { return vm.cid }

func (vm VMImpl) SetProps(props VMProps) error {
	_, err := vm.driver.Execute(
		"modifyvm", vm.cid.AsString(),
		"--name", vm.cid.AsString(),
		"--memory", strconv.Itoa(props.Memory),
		"--cpus", strconv.Itoa(props.CPUs),
		"--paravirtprovider", props.ParavirtProvider,
		"--audio", props.Audio,
	)
	if err != nil {
		return err
	}

	for index, folder := range props.SharedFolders {
		name := fmt.Sprintf("folder-%d", index)

		_, err := vm.driver.Execute(
			"setextradata", vm.cid.AsString(),
			"VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+name, "1",
		)
		if err != nil {
			return err
		}

		_, err = vm.driver.Execute(
			"sharedfolder", "add", vm.cid.AsString(),
			"--name", name,
			"--hostpath", folder.HostPath,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (vm VMImpl) SetMetadata(meta apiv1.VMMeta) error {
	// todo can we do better?
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

func (vm VMImpl) ConfigureNICs(nets Networks, host Host) error {
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

	_, err = vm.driver.Execute("unregistervm", vm.cid.AsString(), "--delete")
	if err != nil {
		return err
	}

	return vm.store.Delete()
}
