package driver

import (
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var (
	execDriverNotReadyErr = regexp.MustCompile("VBoxManage: error: The object is not ready")
	execDriverDevCtlErr   = regexp.MustCompile(`failed to open \/dev\/vboxnetctl`)
	execDriverGenericErr  = regexp.MustCompile("VBoxManage: error:")
)

type ExecDriver struct {
	runner  Runner
	retrier Retrier
	binPath string

	logTag string
	logger boshlog.Logger
}

func NewExecDriver(runner Runner, retrier Retrier, binPath string, logger boshlog.Logger) ExecDriver {
	return ExecDriver{
		runner:  runner,
		retrier: retrier,
		binPath: binPath,

		logTag: "driver.ExecDriver",
		logger: logger,
	}
}

func (d ExecDriver) Execute(args ...string) (string, error) {
	return d.ExecuteComplex(args, ExecuteOpts{})
}

func (d ExecDriver) ExecuteComplex(args []string, opts ExecuteOpts) (string, error) {
	var output string
	var status int

	execFunc := func() error {
		var err error

		output, status, err = d.runner.Execute(d.binPath, args...)
		if err != nil {
			return RetryableErrorImpl{err}
		}

		if status != 0 && execDriverNotReadyErr.MatchString(output) {
			return RetryableErrorImpl{err}
		}

		return nil
	}

	err := d.retrier.Retry(execFunc)
	output = strings.Replace(output, "\r\n", "\n", -1)
	if err != nil {
		return output, err
	}

	var errored bool

	if status != 0 {
		if status == 126 {
			// This exit code happens if VBoxManage is on the PATH,
			// but another executable it tries to execute is missing.
			// This is usually indicative of a corrupted VirtualBox install.
			return output, bosherr.Errorf("Most likely corrupted VirtualBox installation")
		} else {
			errored = !opts.IgnoreNonZeroExitStatus
		}
	} else {
		// Sometimes, VBoxManage fails but doesn't actual return a non-zero exit code.
		if execDriverDevCtlErr.MatchString(output) {
			// This catches an error message that only shows when kernel
			// drivers aren't properly installed.
			return output, bosherr.Errorf("Error message about vboxnetctl")
		}

		if execDriverGenericErr.MatchString(output) {
			d.logger.Debug(d.logTag, "VBoxManage error text found, assuming error.")
			errored = true
		}
	}

	if errored {
		return output, bosherr.Errorf("Error executing command:\nCommand: '%v'\nExit code: %d\nOutput: '%s'", args, status, output)
	}

	return output, nil
}

func (d ExecDriver) IsMissingVMErr(output string) bool {
	return strings.Contains(output, "Could not find a registered machine with UUID")
}
