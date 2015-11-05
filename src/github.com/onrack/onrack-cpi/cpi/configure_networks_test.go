package cpi_test

import (
	"github.com/onrack/onrack-cpi/onrackapi"
	"github.com/onrack/onrack-cpi/cpi"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigureNetworks", func() {
	It("Returns unsupported error", func() {
		instanceId := "instance_id"
		network := onrackapi.Network{}
		err := cpi.ConfigureNetworks(instanceId, network)
		Expect(err).To(HaveOccurred())
	})
})
