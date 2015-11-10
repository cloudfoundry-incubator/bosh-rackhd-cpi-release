package cpi_test

import (
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

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
			apiServerIP := os.Getenv("RACKHD_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())

			config := config.Cpi{ApiServer: apiServerIP}

			var input bosh.MethodArguments
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

			fileMetadataResp := rackhdapi.FileMetadataResponse{}
			err = json.Unmarshal(respBytes, &fileMetadataResp)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileMetadataResp).To(HaveLen(1))

			fileMetadata := fileMetadataResp[0]
			Expect(fileMetadata.Basename).To(Equal(uuid))

			err = rackhdapi.DeleteFile(config, fileMetadata.UUID)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("With invalid CPI v1 input", func() {
		It("Returns an error", func() {
			config := config.Cpi{}

			var input bosh.MethodArguments
			input = append(input, map[string]string{"foo": "bar"})

			uuid, err := cpi.CreateStemcell(config, input)
			Expect(err).To(MatchError("received unexpected type for stemcell image path"))
			Expect(uuid).To(BeEmpty())
		})
	})
})
