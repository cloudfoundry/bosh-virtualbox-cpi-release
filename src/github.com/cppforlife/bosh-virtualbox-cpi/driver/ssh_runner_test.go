package driver_test

import (
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var (
	testSSHRunnerUsername   = os.Getenv("TEST_SSH_RUNNER_USERNAME")
	testSSHRunnerPrivateKey = os.Getenv("TEST_SSH_RUNNER_PRIVATE_KEY")
	testSSHRunnerHost       = os.Getenv("TEST_SSH_RUNNER_HOST")
)

var _ = Describe("SSHRunner", func() {
	BeforeEach(func() {
		if len(testSSHRunnerUsername) == 0 {
			Skip("SSHRunner tests are not configured")
		}
		if testSSHRunnerHost == "" {
			testSSHRunnerHost = "127.0.0.1"
		}
	})

	Describe("HomeDir", func() {
		It("returns proper home directory", func() {
			opts := SSHRunnerOpts{
				Host:       testSSHRunnerHost,
				Username:   testSSHRunnerUsername,
				PrivateKey: testSSHRunnerPrivateKey,
			}
			logger := boshlog.NewLogger(boshlog.LevelNone)
			runner := NewSSHRunner(opts, nil, logger)

			path, err := runner.HomeDir()
			Expect(err).ToNot(HaveOccurred())
			Expect(path).ToNot(BeEmpty())
			Expect(path).ToNot(ContainSubstring("~"))
		})
	})
})
