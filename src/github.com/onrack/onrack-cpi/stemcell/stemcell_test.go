package stemcell_test

import (
	"os"

	"github.com/onrack/onrack-cpi/stemcell"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stemcell", func() {
	Describe("Extract()", func() {
		Context("With a valid stemcell", func() {
			It("returns a file handle to the VMDK contained in a vSphere stemcell", func() {
				stemcell := stemcell.New("../spec_assets/image")
				stemcellHandle, err := stemcell.Extract()
				Expect(stemcellHandle).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())

				var testFile *os.File
				Expect(stemcellHandle).To(BeAssignableToTypeOf(testFile))
				stemcellBytes, err := ioutil.ReadAll(stemcellHandle)
				Expect(err).ToNot(HaveOccurred())
				Expect(stemcellBytes).ToNot(BeEmpty())
			})
		})
		Context("When the stemcell is invalid", func() {
			It("returns an error", func() {
				invalidStemcellArchive, err := ioutil.TempFile("", "invalid-stemcell")
				Expect(err).ToNot(HaveOccurred())
				defer invalidStemcellArchive.Close()
				defer os.Remove(invalidStemcellArchive.Name())

				stemcell := stemcell.New(invalidStemcellArchive.Name())
				stemcellHandle, err := stemcell.Extract()
				Expect(err).To(HaveOccurred())
				Expect(stemcellHandle).To(BeNil())
			})
		})
		Context("When the stemcell does not exist", func() {
			It("returns an error", func() {
				stemcell := stemcell.New("/probably-not-a-file-hopefully")
				stemcellHandle, err := stemcell.Extract()
				Expect(err).To(HaveOccurred())
				Expect(stemcellHandle).To(BeNil())
			})
		})
	})
	Describe("Cleanup()", func() {
		It("cleans up the temp directory and closes the file handle?", func() {
			stemcell := stemcell.New("../spec_assets/image")
			stemcellHandle, err := stemcell.Extract()
			Expect(stemcellHandle).ToNot(BeNil())
			Expect(err).ToNot(HaveOccurred())

			workDir := stemcell.GetWorkDir()
			Expect(workDir).ToNot(BeEmpty())

			oldWorkDirContents, err := ioutil.ReadDir(workDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(oldWorkDirContents).ToNot(BeEmpty())
			Expect(oldWorkDirContents).To(HaveLen(1))

			stemcellFileInfo := oldWorkDirContents[0]
			Expect(stemcellFileInfo.Name()).To(Equal("image-disk1.vmdk"))

			err = stemcell.CleanUp()
			Expect(err).ToNot(HaveOccurred())
			Expect(workDir).ToNot(BeADirectory())
		})
	})
})
