package cpi_test

import (
	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/cpi"
	"github.com/onrack/onrack-cpi/onrackhttp"

	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"io/ioutil"
	"os"
)

var _ = Describe("CreateStemcell", func() {
	Context("With valid CPI v1 input", func() {
		It("Uploads the image from an OpenStack stemcell", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())

			config := config.Cpi{ApiServer: apiServerIP}

			var input bosh.ExternalInput
			input = append(input, "../spec_assets/image")

			uuid, err := cpi.CreateStemcell(config, input)
			Expect(err).ToNot(HaveOccurred())

			Expect(uuid).ToNot(BeEmpty())
			url := fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", config.ApiServer, uuid)
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))

			fileMetadataResp := onrackhttp.FileMetadataResponse{}
			err = json.Unmarshal(respBytes, &fileMetadataResp)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileMetadataResp).To(HaveLen(1))

			fileMetadata := fileMetadataResp[0]
			Expect(fileMetadata.Basename).To(Equal(uuid))

			err = onrackhttp.DeleteFile(config, fileMetadata.UUID)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("With invalid CPI v1 input", func() {
		It("Returns an error", func() {
			config := config.Cpi{}

			var input bosh.ExternalInput
			input = append(input, map[string]string{"foo": "bar"})

			uuid, err := cpi.CreateStemcell(config, input)
			Expect(err).To(MatchError("Received unexpected type for stemcell image path"))
			Expect(uuid).To(BeEmpty())
		})
	})
})
