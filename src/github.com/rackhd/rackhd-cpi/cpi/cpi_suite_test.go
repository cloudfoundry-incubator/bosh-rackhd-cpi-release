package cpi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCpi(t *testing.T) {
	// where did my logs go
	// disable logging
	// log.SetOutput(ioutil.Discard)

	RegisterFailHandler(Fail)
	RunSpecs(t, "CPI Suite")
}
