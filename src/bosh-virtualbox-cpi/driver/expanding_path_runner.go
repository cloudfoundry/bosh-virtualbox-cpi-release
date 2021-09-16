package driver

import (
	"path/filepath"
	"strings"
)

var (
	homeMarker = "~"
)

type ExpandingPathRunner struct {
	other           RawRunner
	resolvedHomeDir string
}

func NewExpandingPathRunner(other RawRunner) *ExpandingPathRunner {
	return &ExpandingPathRunner{other, ""}
}

func (r *ExpandingPathRunner) Execute(path string, args ...string) (string, int, error) {
	args, err := r.expandPaths(args)
	if err != nil {
		return "", 0, err
	}

	return r.other.Execute(path, args...)
}

func (r *ExpandingPathRunner) Upload(srcPath, dstPath string) error {
	srcPath, err := r.expandPath(srcPath)
	if err != nil {
		return err
	}

	dstPath, err = r.expandPath(dstPath)
	if err != nil {
		return err
	}

	return r.other.Upload(srcPath, dstPath)
}

func (r *ExpandingPathRunner) Put(path string, contents []byte) error {
	path, err := r.expandPath(path)
	if err != nil {
		return err
	}
	return r.other.Put(path, contents)
}

func (r *ExpandingPathRunner) Get(path string) ([]byte, error) {
	path, err := r.expandPath(path)
	if err != nil {
		return nil, err
	}
	return r.other.Get(path)
}

func (r *ExpandingPathRunner) expandPaths(args []string) ([]string, error) {
	var expandedArgs []string
	var err error
	for _, arg := range args {
		arg, err = r.expandPath(arg)
		if err != nil {
			return nil, err
		}
		expandedArgs = append(expandedArgs, arg)
	}
	return expandedArgs, nil
}

func (r *ExpandingPathRunner) expandPath(arg string) (string, error) {
	if strings.HasPrefix(arg, homeMarker) {
		homeDir, err := r.homeDir()
		if err != nil {
			return "", err
		}
		arg = filepath.Join(homeDir, strings.TrimPrefix(arg, homeMarker))
	}
	return arg, nil
}

func (r *ExpandingPathRunner) homeDir() (string, error) {
	if len(r.resolvedHomeDir) > 0 {
		return r.resolvedHomeDir, nil
	}

	var err error

	r.resolvedHomeDir, err = r.other.HomeDir()

	return r.resolvedHomeDir, err
}
