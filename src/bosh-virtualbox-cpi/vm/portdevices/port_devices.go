package portdevices

import (
	"fmt"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	"bosh-virtualbox-cpi/driver"
)

var (
	// Covers `storagecontrollername1="SCSI"` and `storagecontrollername1="SCSI Controller"`
	scsiControllerName = regexp.MustCompile(`^storagecontrollername\d+="((?i:SCSI).*)"$`)

	// Covers `storagecontrollername="IDE"` and `storagecontrollername0="IDE Controller"`
	ideControllerName = regexp.MustCompile(`^storagecontrollername\d+="((?i:IDE).*)"$`)

	// Covers `"storagecontrollername0="SATA"` and `storagecontrollername0="SATA Controller"`
	sataControllerName = regexp.MustCompile(`^storagecontrollername\d+="((?i:SATA).*)"$`)

	// Covers `SCSI-1-0"="none"` and `"SCSI Controller-1-0"="none"` (name-port-device)
	portDeviceConfig = regexp.MustCompile(`^"(?:.+)-(\d+)-(\d+)"="none"$`)
)

type PortDevicesOpts struct {
	Controller string
}

type PortDevices struct {
	vmCID apiv1.VMCID
	opts  PortDevicesOpts

	driver driver.Driver
	logger boshlog.Logger
}

func NewPortDevices(vmCID apiv1.VMCID, opts PortDevicesOpts, driver driver.Driver, logger boshlog.Logger) PortDevices {
	return PortDevices{vmCID, opts, driver, logger}
}

func (d PortDevices) CDROM() (CDROM, error) {
	// Always use SCSI for CDROM as IDE controller slots are limited
	name, _, err := d.determineControllerName(scsiControllerName)
	if err != nil {
		return CDROM{}, err
	}

	// todo pick available?
	return CDROM{driver: d.driver, vmCID: d.vmCID, name: name, port: "0", device: "0"}, nil
}

func (d PortDevices) FindAvailable() (PortDevice, error) {
	pds, err := d.availablePDs()
	if err != nil {
		return PortDevice{}, err
	}

	if len(pds) == 0 {
		return PortDevice{}, bosherr.Error("No available controller port & device")
	}

	return pds[0], nil
}

func (d PortDevices) Find(controller, port, device string) (PortDevice, error) {
	var controllerNameMatch *regexp.Regexp

	switch controller {
	case IDEController:
		controllerNameMatch = ideControllerName
	case SATAController:
		controllerNameMatch = sataControllerName
	case SCSIController, "":
		controller = SCSIController
		controllerNameMatch = scsiControllerName
	default:
		panic(fmt.Sprintf("Unexpected storage controller '%s'", controller))
	}

	name, _, err := d.determineControllerName(controllerNameMatch)
	if err != nil {
		return PortDevice{}, err
	}

	return NewPortDevice(d.driver, d.vmCID, controller, name, port, device), nil
}

func (d PortDevices) availablePDs() ([]PortDevice, error) {
	var controllerNameMatch *regexp.Regexp

	switch d.opts.Controller {
	case IDEController:
		controllerNameMatch = ideControllerName
	case SCSIController:
		controllerNameMatch = scsiControllerName
	case SATAController:
		controllerNameMatch = sataControllerName
	default:
		panic(fmt.Sprintf("Unexpected storage controller '%s'", d.opts.Controller))
	}

	name, output, err := d.determineControllerName(controllerNameMatch)
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
			pd := NewPortDevice(d.driver, d.vmCID, d.opts.Controller, name, matches[1], matches[2])
			pds = append(pds, pd)
		}
	}

	return pds, nil
}

func (d PortDevices) determineControllerName(nameMatch *regexp.Regexp) (string, string, error) {
	output, err := d.driver.Execute("showvminfo", d.vmCID.AsString(), "--machinereadable")
	if err != nil {
		return "", output, bosherr.WrapErrorf(err, "Determining controller name")
	}

	for _, line := range strings.Split(output, "\n") {
		matches := nameMatch.FindStringSubmatch(line)
		if len(matches) > 0 {
			if len(matches) != 2 {
				panic(fmt.Sprintf("Internal inconsistency: Expected len(%s matches) == 2", nameMatch))
			}

			d.logger.Debug("vm.PortDevices",
				"Determined controller name '%s' from output '%s'", matches[1], output)

			return matches[1], output, nil
		}
	}

	return "", output, bosherr.Error("Unknown controller name")
}
