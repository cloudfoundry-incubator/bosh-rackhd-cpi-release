package cpi_test

import (
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

var _ = Describe("DeleteStemcell", func() {
	Context("Valid input", func() {
		var c config.Cpi
		var fileName string
		var err error

		BeforeEach(func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())
			c = config.Cpi{ApiServer: apiServer}
			var createInput bosh.MethodArguments
			createInput = append(createInput, "../spec_assets/image")

			fileName, err = cpi.CreateStemcell(c, createInput)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			var deleteInput bosh.MethodArguments
			deleteInput = append(deleteInput, fileName)
			err = cpi.DeleteStemcell(c, deleteInput)
			Expect(err).ToNot(HaveOccurred())

			_, err = rackhdapi.GetFile(c, fileName)
			Expect(err).To(HaveOccurred())
		})

		Context("with valid CPI v1 input", func() {
			It("deletes a previously uploaded stemcell from the rackhd server", func() {
				_, err = rackhdapi.GetFile(c, fileName)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("with invalid CPI v1 input", func() {
		It("returns an error", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())
			c := config.Cpi{ApiServer: apiServer}

			var deleteInput bosh.MethodArguments
			deleteInput = append(deleteInput, map[string]string{"invalid": "true"})
			err = cpi.DeleteStemcell(c, deleteInput)
			Expect(err).To(MatchError("Received unexpected type for stemcell cid"))
		})
	})
})
