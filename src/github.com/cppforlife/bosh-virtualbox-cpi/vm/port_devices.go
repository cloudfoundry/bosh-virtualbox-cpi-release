package vm

import (
	"fmt"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var (
	// Covers `storagecontrollername1="SCSI"` and `storagecontrollername1="SCSI Controller"`
	scsiControllerName = regexp.MustCompile(`^storagecontrollername\d+="((?i:SCSI).*)"$`)
	// Covers `SCSI-1-0"="none"` and `"SCSI Controller-1-0"="none"`
	portDeviceConfig = regexp.MustCompile(`^"(?:.+)-(\d+)-(\d+)"="none"$`)
)

type PortDevices struct {
	driver driver.Driver
	vm     VM
}

func (d PortDevices) FindAvailable() (PortDevice, error) {
	pds, err := d.availablePDs()
	if err != nil {
		return PortDevice{}, err
	}

	if len(pds) == 0 {
		return PortDevice{}, bosherr.Error("No available SCSI Controller port&device")
	}

	return pds[0], nil
}

func (d PortDevices) Find(port, device string) (PortDevice, error) {
	name, _, err := d.determineSCSIControllerName()
	if err != nil {
		return PortDevice{}, err
	}

	return NewPortDevice(d.driver, d.vm, name, port, device), nil
}

func (d PortDevices) availablePDs() ([]PortDevice, error) {
	name, output, err := d.determineSCSIControllerName()
	if err != nil {
		return nil, err
	}

	var pds []PortDevice

	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, fmt.Sprintf("\"%s-", name)) {
			continue
		}

		matches := portDeviceConfig.FindStringSubmatch(line)
		if len(matches) > 0 {
			if len(matches) != 3 {
				panic("Internal inconsistency: Expected len(portDeviceConfig matches) == 3")
			}
			pd := NewPortDevice(d.driver, d.vm, name, matches[1], matches[2])
			pds = append(pds, pd)
		}
	}

	return pds, nil
}

func (d PortDevices) determineSCSIControllerName() (string, string, error) {
	output, err := d.driver.Execute("showvminfo", d.vm.ID(), "--machinereadable")
	if err != nil {
		return "", output, bosherr.WrapErrorf(err, "Determining SCSI controller name")
	}

	for _, line := range strings.Split(output, "\n") {
		matches := scsiControllerName.FindStringSubmatch(line)
		if len(matches) > 0 {
			if len(matches) != 2 {
				panic("Internal inconsistency: Expected len(scsiControllerName matches) == 2")
			}
			return matches[1], output, nil
		}
	}

	return "", output, bosherr.Error("Unknown SCSI controller name")
}
