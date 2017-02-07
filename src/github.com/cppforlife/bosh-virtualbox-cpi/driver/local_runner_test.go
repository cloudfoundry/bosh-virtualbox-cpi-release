package driver_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-virtualbox-cpi/driver"
)

var _ = Describe("LocalRunner", func() {
	Describe("HomeDir", func() {
		It("returns proper home directory", func() {
			logger := boshlog.NewLogger(boshlog.LevelNone)
			cmdRunner := boshsys.NewExecCmdRunner(logger)
			runner := NewLocalRunner(nil, cmdRunner, logger)

			path, err := runner.HomeDir()
			Expect(err).ToNot(HaveOccurred())
			Expect(path).ToNot(BeEmpty())
			Expect(path).ToNot(ContainSubstring("~"))
		})
	})
})
