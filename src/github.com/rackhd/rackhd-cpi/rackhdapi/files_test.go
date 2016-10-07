package rackhdapi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/nu7hatch/gouuid"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Files", func() {
	Describe("uploading to then deleting from the RackHD API", func() {
		It("allows files to be uploaded and deleted", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			c := config.Cpi{ApiServer: apiServer}
			dummyStr := "Some ice cold file"
			dummyFile := strings.NewReader(dummyStr)

			uuid, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			baseName := uuid.String()

			url := fmt.Sprintf("%s/api/2.0/files/%s/metadata", c.ApiServer, baseName)
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))

			rackhdUUID, err := rackhdapi.UploadFile(c, baseName, dummyFile, int64(len(dummyStr)))
			Expect(err).ToNot(HaveOccurred())
			Expect(rackhdUUID).ToNot(BeEmpty())

			getResp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(getResp.Body)
			Expect(err).ToNot(HaveOccurred())

			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(200))

			fileMetadataResp := rackhdapi.FileMetadataResponse{}
			fmt.Printf("resp: %+v", string(respBytes))
			err = json.Unmarshal(respBytes, &fileMetadataResp)
			Expect(err).ToNot(HaveOccurred())
			//Expect(fileMetadataResp).To(HaveLen(1))
			Expect(fileMetadataResp.Basename).ToNot(BeEmpty())
			Expect(fileMetadataResp.Md5).ToNot(BeEmpty())
			Expect(fileMetadataResp.Sha256).ToNot(BeEmpty())
			Expect(fileMetadataResp.UUID).ToNot(BeEmpty())

			fileMetadata := fileMetadataResp
			Expect(fileMetadata.Basename).To(Equal(baseName))

			err = rackhdapi.DeleteFile(c, baseName)
			Expect(err).ToNot(HaveOccurred())

			resp, err = http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})
	})
})
