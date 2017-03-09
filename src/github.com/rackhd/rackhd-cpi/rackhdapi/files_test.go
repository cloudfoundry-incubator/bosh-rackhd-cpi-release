package rackhdapi_test

import (
	"strings"

	"github.com/nu7hatch/gouuid"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Files", func() {
	Describe("uploading to then deleting from the RackHD API", func() {
		var baseName string
		var uploadResponse models.FileUploadResponse
		var c config.Cpi

		BeforeEach(func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			c = config.Cpi{ApiServer: apiServer}
			dummyStr := "Some ice cold file"
			dummyFile := strings.NewReader(dummyStr)

			uuid, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())

			baseName = uuid.String()
			_, err = rackhdapi.GetFile(c, baseName)
			Expect(err).To(HaveOccurred())

			uploadResponse, err = rackhdapi.UploadFile(c, baseName, dummyFile, int64(len(dummyStr)))
			Expect(err).ToNot(HaveOccurred())
			Expect(uploadResponse.Name).To(Equal(baseName))
		})

		AfterEach(func() {
			err := rackhdapi.DeleteFile(c, uploadResponse.Name)
			Expect(err).ToNot(HaveOccurred())

			_, err = rackhdapi.GetFile(c, uploadResponse.Name)
			Expect(err).To(HaveOccurred())
		})

		It("allows files to be uploaded and deleted", func() {
			getFile, err := rackhdapi.GetFile(c, uploadResponse.UUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(getFile)).To(Equal("Some ice cold file"))
		})
	})
})
