package cpi_test

import (
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateStemcell", func() {
	Context("With valid CPI v1 input", func() {
		var fileUUID string
		var c config.Cpi
		var input bosh.MethodArguments
		var err error

		BeforeEach(func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())
			c = config.Cpi{ApiServer: apiServer}
			input = append(input, "../spec_assets/image")
		})

		AfterEach(func() {
			err := rackhdapi.DeleteFile(c, fileUUID)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Uploads the image from an OpenStack stemcell", func() {
			fileUUID, err = cpi.CreateStemcell(c, input)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileUUID).ToNot(BeEmpty())

			_, err = rackhdapi.GetFile(c, fileUUID)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("With invalid CPI v1 input", func() {
		It("Returns an error", func() {
			config := config.Cpi{}
			var input bosh.MethodArguments
			input = append(input, map[string]string{"foo": "bar"})

			fileUUID, err := cpi.CreateStemcell(config, input)
			Expect(err).To(MatchError("received unexpected type for stemcell image path"))
			Expect(fileUUID).To(BeEmpty())
		})
	})
})
