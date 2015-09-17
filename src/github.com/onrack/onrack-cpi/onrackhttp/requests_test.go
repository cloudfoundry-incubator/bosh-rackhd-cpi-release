package onrackhttp_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Requests", func() {
	Describe("uploading to then deleting from the OnRack API", func() {
		It("allows files to be uploaded and deleted", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}
			dummyStr := "Some ice cold file"
			dummyFile := strings.NewReader(dummyStr)

			uuid, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())

			url := fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", c.ApiServer, uuid.String())
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))

			onrackUUID, err := onrackhttp.UploadFile(c, uuid.String(), dummyFile, int64(len(dummyStr)))
			Expect(err).ToNot(HaveOccurred())
			Expect(onrackUUID).ToNot(BeEmpty())

			url = fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", c.ApiServer, uuid.String())
			getResp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(getResp.Body)
			Expect(err).ToNot(HaveOccurred())

			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(200))

			fileMetadataResp := onrackhttp.FileMetadataResponse{}
			err = json.Unmarshal(respBytes, &fileMetadataResp)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileMetadataResp).To(HaveLen(1))

			fileMetadata := fileMetadataResp[0]
			Expect(fileMetadata.Basename).To(Equal(uuid.String()))

			err = onrackhttp.DeleteFile(c, onrackUUID)
			Expect(err).ToNot(HaveOccurred())

			url = fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", c.ApiServer, uuid.String())
			resp, err = http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})
	})
})
