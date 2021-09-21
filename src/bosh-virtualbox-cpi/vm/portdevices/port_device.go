package portdevices

import (
	"fmt"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-virtualbox-cpi/driver"
)

var (
	// Covers `"SATA-ImageUUID-0-0"="a840e5e0-947c-4e63-ac2f-678a86d13980"`
	portImagewUUIDConfig = regexp.MustCompile(`^"(\w+)-ImageUUID-(\d+)-(\d+)"="(.+)"$`)
)

type PortDevice struct {
	driver driver.Driver
	vmCID  apiv1.VMCID

	controller string // e.g. scsi, ide, sata
	name       string // e.g. IDE, SCSI, AHCI Controller

	port   string
	device string
}

func NewPortDevice(driver driver.Driver, vmCID apiv1.VMCID, controller, name, port, device string) PortDevice {
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
		vmCID:  vmCID,

		controller: controller,
		name:       name,

		port:   port,
		device: device,
	}
}

func (d PortDevice) Controller() string { return d.controller }

func (d PortDevice) Port() string   { return d.port }
func (d PortDevice) Device() string { return d.device }

func (d PortDevice) Hint() apiv1.DiskHint {
	switch d.controller {
	case IDEController:
		switch {
		case d.port == "0": // Assume system disk is 0
			return apiv1.NewDiskHintFromString(d.device)
		default:
			// todo unsafe disk selection!
			// todo does not work on reboot
			// Ideally will specify scsi_host_no in addition to scsi_id
			// (https://www.ibm.com/support/knowledgecenter/linuxonibm/com.ibm.linux.z.lgdd/lgdd_t_fcp_wrk_uinfo.html)
			return apiv1.NewDiskHintFromMap(map[string]interface{}{"id": "1ATA"})
		}

	case SCSIController:
		// Assumes that all ports are connected to the root disk device
		// given how current bosh-agent tries to find disks
		// todo ideally specify port & device (unrelated to root disk)
		return apiv1.NewDiskHintFromString(d.port)

	case SATAController:
		// First section of imageUUID appears under /dev/disk/by-id
		// example ata-VBOX_HARDDISK_VBb1ade788-61d20269
		// b1ade788 would be first section of imageUUID
		imageUUID, err := d.imageUUID()
		if err != nil {
			panic(bosherr.WrapError(err, "Disk hint"))
		}

		prefix := strings.Split(imageUUID, "-")[0]

		// DiskHint used by: SCSIIDDevicePathResolver
		// "id" gets mapped into DeviceID by EphemeralDiskSettings()
		return apiv1.NewDiskHintFromMap(map[string]interface{}{"id": prefix + "*"})

	default:
		panic(fmt.Sprintf("Unexpected storage controller '%s'", d.name))
	}
}

func (d PortDevice) Attach(path string) error {
	_, err := d.driver.Execute(
		"storageattach", d.vmCID.AsString(),
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
		"storageattach", d.vmCID.AsString(),
		"--storagectl", d.name,
		"--port", d.port,
		"--device", d.device,
		"--type", "hdd",
		"--medium", "none", // removes
	)
	return err
}

func (d PortDevice) imageUUID() (string, error) {
	output, err := d.driver.Execute("showvminfo", d.vmCID.AsString(), "--machinereadable")
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Determining imageUUID")
	}

	for _, line := range strings.Split(output, "\n") {
		matches := portImagewUUIDConfig.FindStringSubmatch(line)
		if len(matches) > 0 {
			if len(matches) != 5 {
				return "", fmt.Errorf("Internal inconsistency: Expected len(%d matches) == 5", len(matches))
			}

			if matches[1] == d.name && matches[2] == d.port && matches[3] == d.device {
				return matches[4], nil
			}
		}
	}
	return "", bosherr.Errorf("Failed to deterime imageUUID for PortDevice %s %s-%s", d.name, d.port, d.device)
}
