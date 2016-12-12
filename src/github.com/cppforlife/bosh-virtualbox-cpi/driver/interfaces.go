package driver

type ExecuteOpts struct {
	IgnoreNonZeroExitStatus bool
}

type Driver interface {
	Execute(args ...string) (string, error)
	ExecuteComplex(args []string, opts ExecuteOpts) (string, error)
}

var _ Driver = ExecDriver{}

type Runner interface {
	Execute(path string, args ...string) (string, int, error)
	Upload(srcDir, dstDir string) error
	Put(path string, contents []byte) error
	Get(path string) ([]byte, error)
}

var _ Runner = LocalRunner{}
