package driver

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"golang.org/x/crypto/ssh"
)

type SSHRunner struct {
	opts SSHRunnerOpts

	fs boshsys.FileSystem

	logTag string
	logger boshlog.Logger

	existingClient *ssh.Client
}

type SSHRunnerOpts struct {
	Host       string
	Username   string
	PrivateKey string
}

func NewSSHRunner(opts SSHRunnerOpts, fs boshsys.FileSystem, logger boshlog.Logger) *SSHRunner {
	return &SSHRunner{opts, fs, "driver.SSHRunner", logger, nil}
}

func (r *SSHRunner) HomeDir() (string, error) {
	output, _, err := r.execute("sh -c 'USER= HOME= eval echo ~`whoami`'")
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(output, "~") {
		return "", bosherr.Errorf("Failed to expand path '%s'", output)
	}

	return strings.TrimSpace(output), nil
}

func (r *SSHRunner) Execute(path string, args ...string) (string, int, error) {
	return r.execute(r.shCmd(path, args, ""))
}

func (r *SSHRunner) execute(cmd string) (string, int, error) {
	r.logger.Debug(r.logTag, "Execute '%s'", cmd)

	sess, err := r.session()
	if err != nil {
		return "", 0, err
	}

	defer sess.Close()

	var stderr, stdout bytes.Buffer
	sess.Stdout = &stdout
	sess.Stderr = &stderr

	err = sess.Run(cmd)
	output := stdout.String() + "\n" + stderr.String()

	if err == nil {
		return output, 0, nil
	}

	switch typedErr := err.(type) {
	case *ssh.ExitMissingError:
		return output, 0, bosherr.WrapError(typedErr, "Missing exit info")
	case *ssh.ExitError:
		status := typedErr.Waitmsg.ExitStatus()
		return output, status, bosherr.WrapErrorf(typedErr, "Exit (Output: '%s')", output)
	default:
		return output, 0, bosherr.WrapErrorf(typedErr, "Unknown SSH error (Output: '%s')", output)
	}
}

func (r *SSHRunner) Upload(srcPath, dstPath string) error {
	r.logger.Debug(r.logTag, "Upload from '%s' to '%s'", srcPath, dstPath)

	file, err := r.fs.OpenFile(srcPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Opening source file for upload")
	}

	defer file.Close()

	return r.putFromReader(dstPath, file)
}

func (r *SSHRunner) Put(path string, contents []byte) error {
	r.logger.Debug(r.logTag, "Put to '%s' %d ", path, len(contents))
	return r.putFromReader(path, bytes.NewBuffer(contents))
}

func (r *SSHRunner) putFromReader(path string, in io.Reader) error {
	sess, err := r.session()
	if err != nil {
		return err
	}

	defer sess.Close()

	sess.Stdin = in

	err = sess.Run(r.shCmd("cat", nil, path))
	if err != nil {
		return bosherr.WrapError(err, "Putting file")
	}

	return nil
}

func (r *SSHRunner) Get(path string) ([]byte, error) {
	r.logger.Debug(r.logTag, "Get '%s'", path)

	sess, err := r.session()
	if err != nil {
		return nil, err
	}

	defer sess.Close()

	var stdout bytes.Buffer
	sess.Stdout = &stdout

	err = sess.Run(r.shCmd("cat", []string{path}, ""))
	if err != nil {
		return nil, bosherr.WrapError(err, "Getting file")
	}

	return stdout.Bytes(), err
}

func (r *SSHRunner) session() (*ssh.Session, error) {
	client, err := r.client()
	if err != nil {
		return nil, err
	}

	sess, err := client.NewSession()
	if err != nil {
		return nil, bosherr.WrapError(err, "Opening SSH session")
	}

	return sess, nil
}

func (r *SSHRunner) client() (*ssh.Client, error) {
	if r.existingClient != nil {
		return r.existingClient, nil
	}

	keySigner, err := ssh.ParsePrivateKey([]byte(r.opts.PrivateKey))
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing private key")
	}

	config := &ssh.ClientConfig{
		User:            r.opts.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(keySigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	r.existingClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", r.opts.Host), config)
	if err != nil {
		return nil, bosherr.WrapError(err, "Connecting via SSH")
	}

	return r.existingClient, nil
}

func (r SSHRunner) shCmd(path string, args []string, stdoutPath string) string {
	escapedCmd := r.shellJoin(append([]string{path}, args...))
	stdoutRedir := ""

	if len(stdoutPath) > 0 {
		stdoutRedir = "> " + r.shellEscape(stdoutPath)
	}

	envPath := "PATH=$PATH:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

	return fmt.Sprintf(`sh -c "%s %s %s"`, envPath, escapedCmd, stdoutRedir)
}

func (r SSHRunner) shellJoin(args []string) string {
	var escapedArgs []string
	for _, arg := range args {
		escapedArgs = append(escapedArgs, r.shellEscape(arg))
	}
	return strings.Join(escapedArgs, " ")
}

var (
	shellEscape   = regexp.MustCompile("([^A-Za-z0-9_\\-.,:\\/@\\n])")
	shellEscapeNl = regexp.MustCompile("\n")
)

// http://ruby-doc.org/stdlib-2.0.0/libdoc/shellwords/rdoc/Shellwords.html#method-c-shelljoin
func (SSHRunner) shellEscape(arg string) string {
	if len(arg) == 0 {
		return "''"
	}

	arg = shellEscape.ReplaceAllString(arg, "\\\\$1")
	arg = shellEscapeNl.ReplaceAllString(arg, "'\n'")

	return arg
}
