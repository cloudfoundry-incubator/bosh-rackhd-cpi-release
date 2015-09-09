package cpi_test

import (
	"github.com/onrack/onrack-cpi/cpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"encoding/json"
	"net/http"

	"fmt"
	"io/ioutil"
)

var _ = Describe("CreateStemcell", func() {
	It("Extracts and uploads a VMDK from a vSphere stemcell", func() {
		var config cpi.Config

		//TODO load ip to use from environment variable
		configStr := []byte(`{"apiserver":"127.0.0.1","agent":{"mbus":"127.0.0.1","ntp":"127.0.0.1"}}`)
		err := json.Unmarshal(configStr, &config)
		Expect(err).ToNot(HaveOccurred())

		var input cpi.ExternalInput
		input = append(input, "stemcell.tgz")

		uuid, err := cpi.CreateStemcell(config, input)
		Expect(err).ToNot(HaveOccurred())

		Expect(uuid).ToNot(BeEmpty())
		url := fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", config.ApiServer, uuid)
		resp, err := http.Get(url)
		Expect(err).ToNot(HaveOccurred())

		respBytes, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		defer resp.Body.Close()

		fileMetadataResp := cpi.FileMetadataResponse{}
		err = json.Unmarshal(respBytes, &fileMetadataResp)
		Expect(err).ToNot(HaveOccurred())
		Expect(fileMetadataResp).To(HaveLen(1))

		fileMetadata := fileMetadataResp[0]
		Expect(fileMetadata.Basename).To(Equal(uuid))
	})

})



