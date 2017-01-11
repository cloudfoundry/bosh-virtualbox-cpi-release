package portdevices

import (
	"fmt"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

type PortDevice struct {
	driver driver.Driver
	vmID   string

	controller string // e.g. scsi, ide
	name       string // e.g. IDE, SCSI Controller

	port   string
	device string
}

func NewPortDevice(driver driver.Driver, vmID, controller, name, port, device string) PortDevice {
	if len(vmID) == 0 {
		panic("Internal inconsistency: PD's vmID must not be empty")
	}
	if len(controller) == 0 {
		panic("Internal inconsistency: PD's controller must not be empty")
	}
	if len(name) == 0 {
		panic("Internal inconsistency: PD's name must not be empty")
	}
	if len(port) == 0 {
		panic("Internal inconsistency: PD's port must not be empty")
	}
	if len(device) == 0 {
		panic("Internal inconsistency: PD's device must not be empty")
	}
	return PortDevice{
		driver: driver,
		vmID:   vmID,

		controller: controller,
		name:       name,

		port:   port,
		device: device,
	}
}

func (d PortDevice) Controller() string { return d.controller }

func (d PortDevice) Port() string   { return d.port }
func (d PortDevice) Device() string { return d.device }

func (d PortDevice) Hint() interface{} {
	switch d.controller {
	case IDEController:
		switch {
		case d.port == "0": // Assume system disk is 0
			return d.device
		default:
			// todo unsafe disk selection!
			// todo does not work on reboot
			// Ideally will specify scsi_host_no in addition to scsi_id
			// (https://www.ibm.com/support/knowledgecenter/linuxonibm/com.ibm.linux.z.lgdd/lgdd_t_fcp_wrk_uinfo.html)
			return map[string]string{"id": "1ATA"}
		}

	case SCSIController:
		// Assumes that all ports are connected to the root disk device
		// given how current bosh-agent tries to find disks
		// todo ideally specify port & device (unrelated to root disk)
		return d.port

	default:
		panic(fmt.Sprintf("Unexpected storage controller '%s'", d.name))
	}
}

func (d PortDevice) Attach(path string) error {
	_, err := d.driver.Execute(
		"storageattach", d.vmID,
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
		"storageattach", d.vmID,
		"--storagectl", d.name,
		"--port", d.port,
		"--device", d.device,
		"--type", "hdd",
		"--medium", "none", // removes
	)
	return err
}
