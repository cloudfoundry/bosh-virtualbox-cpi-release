package driver

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type LocalRunner struct {
	fs        boshsys.FileSystem
	cmdRunner boshsys.CmdRunner
}

func NewLocalRunner(fs boshsys.FileSystem, cmdRunner boshsys.CmdRunner) LocalRunner {
	return LocalRunner{fs, cmdRunner}
}

func (r LocalRunner) Execute(path string, args ...string) (string, int, error) {
	stdout, stderr, status, err := r.cmdRunner.RunCommand(path, args...)
	return stdout + "\n" + stderr, status, err
}

func (r LocalRunner) Upload(srcDir, dstDir string) error {
	return r.fs.Rename(srcDir, dstDir)
}

func (r LocalRunner) Put(path string, contents []byte) error {
	return r.fs.WriteFile(path, contents)
}

func (r LocalRunner) Get(path string) ([]byte, error) {
	return r.fs.ReadFile(path)
}
