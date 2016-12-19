package vm

import (
	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type PortDevice struct {
	driver driver.Driver
	vm     VM

	name   string // e.g. SCSI Controller
	port   string
	device string
}

func NewPortDevice(driver driver.Driver, vm VM, name, port, device string) PortDevice {
	if len(name) == 0 {
		panic("Internal inconsistency: PD's name must not be empty")
	}
	if len(port) == 0 {
		panic("Internal inconsistency: PD's port must not be empty")
	}
	if len(device) == 0 {
		panic("Internal inconsistency: PD's device must not be empty")
	}
	return PortDevice{driver: driver, vm: vm, name: name, port: port, device: device}
}

func (d PortDevice) Attach(path string) error {
	_, err := d.driver.Execute(
		"storageattach", d.vm.ID(),
		"--storagectl", d.name,
		"--port", d.port,
		"--device", d.device,
		"--type", "hdd",
		"--medium", path,
		"--mtype", "normal",
	)
	return err
}

func (d PortDevice) Detach() error {
	_, err := d.driver.Execute(
		"storageattach", d.vm.ID(),
		"--storagectl", d.name,
		"--port", d.port,
		"--device", d.device,
		"--type", "hdd",
		"--medium", "none", // removes
	)
	return err
}
