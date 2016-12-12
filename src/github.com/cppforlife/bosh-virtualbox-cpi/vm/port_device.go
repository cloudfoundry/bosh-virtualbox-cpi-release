package vm

import (
	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type PortDevice struct {
	driver driver.Driver
	vm     VM

	port   string
	device string
}

func (d PortDevice) Attach(path string) error {
	_, err := d.driver.Execute(
		"storageattach", d.vm.ID(),
		"--storagectl", "SCSI Controller",
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
		"--storagectl", "SCSI Controller",
		"--port", d.port,
		"--device", d.device,
		"--type", "hdd",
		"--medium", "none", // removes
	)
	return err
}
