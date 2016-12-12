package vm

import (
	"regexp"
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var (
	portDeviceConfig = regexp.MustCompile(`^"SCSI Controller-(\d+)-(\d+)"="none"$`)
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

func (d PortDevices) Find(port, device string) PortDevice {
	return PortDevice{d.driver, d.vm, port, device}
}

func (d PortDevices) availablePDs() ([]PortDevice, error) {
	time.Sleep(5 * time.Second) // todo remove

	output, err := d.driver.Execute("showvminfo", d.vm.ID(), "--machinereadable")
	if err != nil {
		return nil, err
	}

	var pds []PortDevice

	for _, line := range strings.Split(output, "\n") {
		matches := portDeviceConfig.FindStringSubmatch(line)
		if len(matches) == 3 {
			pd := PortDevice{driver: d.driver, vm: d.vm, port: matches[1], device: matches[2]}
			pds = append(pds, pd)
		}
	}

	return pds, nil
}
