package driver_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "bosh-virtualbox-cpi/driver"
)

var _ = Describe("LocalRunner", func() {
	var (
		logger    boshlog.Logger
		cmdRunner boshsys.CmdRunner
		runner    LocalRunner
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		cmdRunner = boshsys.NewExecCmdRunner(logger)
		runner = NewLocalRunner(nil, cmdRunner, logger)
	})

	Context("HomeDir", func() {
		It("returns proper home directory", func() {

			path, err := runner.HomeDir()
			Expect(err).ToNot(HaveOccurred())
			Expect(path).ToNot(BeEmpty())
			Expect(path).ToNot(ContainSubstring("~"))
		})
	})
})
