package cpi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateStemcell", func() {
	Context("With valid CPI v1 input", func() {
		It("Uploads the image from an OpenStack stemcell", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			config := config.Cpi{ApiServer: apiServer}
			var input bosh.MethodArguments
			input = append(input, "../spec_assets/image")

			uuid, err := cpi.CreateStemcell(config, input)
			Expect(err).ToNot(HaveOccurred())
			Expect(uuid).ToNot(BeEmpty())

			url := fmt.Sprintf("%s/api/2.0/files/%s/metadata", config.ApiServer, uuid)
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()

			fileMetadataResp := models.FileMetadataResponse{}
			err = json.Unmarshal(respBytes, &fileMetadataResp)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileMetadataResp.Basename).To(Equal(uuid))

			err = rackhdapi.DeleteFile(config, fileMetadataResp.Basename)
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
