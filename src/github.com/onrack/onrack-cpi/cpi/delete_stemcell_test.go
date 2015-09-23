package cpi_test

import (
	"fmt"
	"net/http"
	"os"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/cpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteStemcell", func() {
	Context("with valid CPI v1 input", func() {
		It("deletes a previously uploaded stemcell from the onrack server", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())

			config := config.Cpi{ApiServer: apiServerIP}

			var createInput bosh.ExternalInput
			createInput = append(createInput, "../spec_assets/image")

			baseName, err := cpi.CreateStemcell(config, createInput)
			Expect(err).ToNot(HaveOccurred())

			var deleteInput bosh.ExternalInput
			deleteInput = append(deleteInput, baseName)
			err = cpi.DeleteStemcell(config, deleteInput)
			Expect(err).ToNot(HaveOccurred())

			url := fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", config.ApiServer, baseName)
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})
	})

	Context("with invalid CPI v1 input", func() {
		It("returns an error", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())

			config := config.Cpi{ApiServer: apiServerIP}

			var deleteInput bosh.ExternalInput
			deleteInput = append(deleteInput, map[string]string{"invalid": "true"})
			err := cpi.DeleteStemcell(config, deleteInput)
			Expect(err).To(MatchError("Received unexpected type for stemcell cid"))
		})
	})
})
