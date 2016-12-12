package cpi_test

import (
  "github.com/rackhd/rackhd-cpi/bosh"
  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/cpi"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
)

var _ = Describe("ConfigureNetworks", func() {
  It("Returns unsupported error", func() {
    err := cpi.ConfigureNetworks(config.Cpi{}, bosh.MethodArguments{})
    Expect(err).To(HaveOccurred())
  })
})
