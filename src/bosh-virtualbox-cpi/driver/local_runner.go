package driver

import (
	"os/user"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type LocalRunner struct {
	fs        boshsys.FileSystem
	cmdRunner boshsys.CmdRunner

	logTag string
	logger boshlog.Logger
}

func NewLocalRunner(fs boshsys.FileSystem, cmdRunner boshsys.CmdRunner, logger boshlog.Logger) LocalRunner {
	return LocalRunner{fs, cmdRunner, "driver.LocalRunner", logger}
}

func (r LocalRunner) HomeDir() (string, error) {
	// todo use fs?
	output, _, err := r.Execute("sh", "-c", "eval echo ~`whoami`")
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(output, "~") {
		return "", bosherr.Errorf("Failed to expand path '%s'", output)
	}

	return strings.TrimSpace(output), nil
}

func (r LocalRunner) Execute(path string, args ...string) (string, int, error) {
	r.logger.Debug(r.logTag, "Execute '%s %s'", path, strings.Join(args, "' '"))

	current_user, userErr := user.Current()
	if userErr != nil {
		return "", -1, userErr
	}

	cmd := boshsys.Command{
		Name: path,
		Args: args,
		Env: map[string]string{
			"PATH":    "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin",
			"LOGNAME": current_user.Username,
			"USER":    current_user.Username,
		},
	}

	stdout, stderr, status, err := r.cmdRunner.RunComplexCommand(cmd)
	return stdout + "\n" + stderr, status, err
}

func (r LocalRunner) Upload(srcPath, dstPath string) error {
	r.logger.Debug(r.logTag, "Upload from '%s' to '%s'", srcPath, dstPath)
	return boshfu.NewFileMover(r.fs).Move(srcPath, dstPath)
}

func (r LocalRunner) Put(path string, contents []byte) error {
	r.logger.Debug(r.logTag, "Put into '%s' %d contents", path, len(contents))
	return r.fs.WriteFile(path, contents)
}

func (r LocalRunner) Get(path string) ([]byte, error) {
	r.logger.Debug(r.logTag, "Get '%s'", path)
	return r.fs.ReadFile(path)
}
