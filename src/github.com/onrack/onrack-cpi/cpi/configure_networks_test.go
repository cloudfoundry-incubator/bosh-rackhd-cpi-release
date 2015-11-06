package cpi_test

import (
	"github.com/onrack/onrack-cpi/cpi"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/bosh"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureNetworks", func() {
	It("Returns unsupported error", func() {
		err := cpi.ConfigureNetworks(config.Cpi{}, bosh.MethodArguments{})
		Expect(err).To(HaveOccurred())
	})
})
