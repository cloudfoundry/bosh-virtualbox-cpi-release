package cpi_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCpi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cpi Suite")
}
