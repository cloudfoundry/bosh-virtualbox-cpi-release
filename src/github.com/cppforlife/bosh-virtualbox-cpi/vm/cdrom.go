package vm

import (
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var (
	// Covers `storagecontrollername="IDE"` and `storagecontrollername0="IDE Controller"`
	ideControllerName = regexp.MustCompile(`^storagecontrollername\d+="((?i:IDE).*)"$`)
)

type CDROM struct {
	driver  driver.Driver
	isoPath string
}

func (cd CDROM) Mount(vm VM) error {
	name, err := cd.determineIDEControllerName(vm)
	if err != nil {
		return err
	}

	_, err = cd.driver.Execute(
		"storageattach", vm.ID(),
		"--storagectl", name,
		"--port", "1",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", cd.isoPath,
	)
	return err
}

func (cd CDROM) Unmount(vm VM) error {
	name, err := cd.determineIDEControllerName(vm)
	if err != nil {
		return err
	}

	_, err = cd.driver.Execute(
		"storageattach", vm.ID(),
		"--storagectl", name,
		"--port", "1",
		"--device", "0",
		"--type", "dvddrive",
		// 'emptydrive' removes medium from the drive; 'none' removes the device
		"--medium", "emptydrive",
	)
	return err
}

func (cd CDROM) determineIDEControllerName(vm VM) (string, error) {
	output, err := cd.driver.Execute("showvminfo", vm.ID(), "--machinereadable")
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Determining IDE controller name")
	}

	for _, line := range strings.Split(output, "\n") {
		matches := ideControllerName.FindStringSubmatch(line)
		if len(matches) > 0 {
			if len(matches) != 2 {
				panic("Internal inconsistency: Expected len(ideControllerName matches) == 2")
			}
			return matches[1], nil
		}
	}

	return "", bosherr.Error("Unknown IDE controller name")
}
