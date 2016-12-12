package vm

import (
	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type CDROM struct {
	driver  driver.Driver
	isoPath string
}

func (cd CDROM) Mount(vm VM) error {
	_, err := cd.driver.Execute(
		"storageattach", vm.ID(),
		"--storagectl", "IDE Controller",
		"--port", "1",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", cd.isoPath,
	)
	return err
}

func (cd CDROM) Unmount(vm VM) error {
	_, err := cd.driver.Execute(
		"storageattach", vm.ID(),
		"--storagectl", "IDE Controller",
		"--port", "1",
		"--device", "0",
		"--type", "dvddrive",
		// 'emptydrive' removes medium from the drive; 'none' removes the device
		"--medium", "emptydrive",
	)
	return err
}
