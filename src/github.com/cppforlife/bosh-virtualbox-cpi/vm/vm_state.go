package vm

import (
	"regexp"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var (
	vmStarted           = regexp.MustCompile(`VM ".+?" has been successfully started`)
	vmStateMatch        = regexp.MustCompile(`VMState="(.+?)"`)
	vmStateInaccessible = regexp.MustCompile(`name="<inaccessible>"`)
)

func (vm VMImpl) Exists() (bool, error) {
	output, err := vm.driver.Execute("showvminfo", vm.id, "--machinereadable")
	if err != nil {
		if vm.driver.IsMissingVMErr(output) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (vm VMImpl) Start(gui bool) error {
	mode := "headless"
	if gui {
		mode = "gui"
	}

	output, err := vm.driver.ExecuteComplex(
		[]string{"startvm", vm.id, "--type", mode},
		driver.ExecuteOpts{IgnoreNonZeroExitStatus: true},
	)
	if err != nil && !vmStarted.MatchString(output) {
		return err
	}

	return nil
}

func (vm VMImpl) Reboot() error {
	err := vm.HaltIfRunning()
	if err != nil {
		return err
	}

	return vm.Start(false) // todo find out previous state
}

func (vm VMImpl) HaltIfRunning() error {
	running, err := vm.IsRunning()
	if err != nil {
		return err
	}

	if running {
		_, err = vm.driver.Execute("controlvm", vm.id, "poweroff")
	}

	return err
}

func (vm VMImpl) IsRunning() (bool, error) {
	state, err := vm.State()
	if err != nil {
		return false, err
	}

	return state == "running", nil
}

func (vm VMImpl) State() (string, error) {
	output, err := vm.driver.Execute("showvminfo", vm.id, "--machinereadable")
	if err != nil {
		return "", err
	}

	if vmStateInaccessible.MatchString(output) {
		return "inaccessible", nil
	}

	matches := vmStateMatch.FindStringSubmatch(output)
	if len(matches) == 2 {
		return matches[1], nil
	}

	return "", bosherr.Errorf("Unknown VM state:\nOutput: '%s'", output)
}
