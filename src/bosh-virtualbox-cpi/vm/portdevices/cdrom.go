package portdevices

import (
	apiv1 "github.com/cloudfoundry/bosh-cpi-go/apiv1"

	"bosh-virtualbox-cpi/driver"
)

type CDROM struct {
	driver driver.Driver
	vmCID  apiv1.VMCID

	name string // e.g. IDE Controller

	port   string
	device string
}

func (cd CDROM) Mount(isoPath string) error {
	_, err := cd.driver.Execute(
		"storageattach", cd.vmCID.AsString(),
		"--storagectl", cd.name,
		"--port", cd.port,
		"--device", cd.device,
		"--type", "dvddrive",
		"--medium", isoPath,
	)
	return err
}

func (cd CDROM) Unmount() error {
	_, err := cd.driver.Execute(
		"storageattach", cd.vmCID.AsString(),
		"--storagectl", cd.name,
		"--port", cd.port,
		"--device", cd.device,
		"--type", "dvddrive",
		// 'emptydrive' removes medium from the drive; 'none' removes the device
		"--medium", "emptydrive",
	)
	return err
}
