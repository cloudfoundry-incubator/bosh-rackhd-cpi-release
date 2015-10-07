package onrackapi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Files", func() {
	Describe("uploading to then deleting from the OnRack API", func() {
		It("allows files to be uploaded and deleted", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}
			dummyStr := "Some ice cold file"
			dummyFile := strings.NewReader(dummyStr)

			uuid, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			baseName := uuid.String()

			url := fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", c.ApiServer, baseName)
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))

			onrackUUID, err := onrackapi.UploadFile(c, baseName, dummyFile, int64(len(dummyStr)))
			Expect(err).ToNot(HaveOccurred())
			Expect(onrackUUID).ToNot(BeEmpty())

			getResp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(getResp.Body)
			Expect(err).ToNot(HaveOccurred())

			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(200))

			fileMetadataResp := onrackapi.FileMetadataResponse{}
			err = json.Unmarshal(respBytes, &fileMetadataResp)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileMetadataResp).To(HaveLen(1))

			fileMetadata := fileMetadataResp[0]
			Expect(fileMetadata.Basename).To(Equal(baseName))

			err = onrackapi.DeleteFile(c, baseName)
			Expect(err).ToNot(HaveOccurred())

			resp, err = http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})
	})
})
