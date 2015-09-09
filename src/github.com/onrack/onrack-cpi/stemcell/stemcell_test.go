package stemcell_test

import (
	"github.com/onrack/onrack-cpi/stemcell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stemcell", func() {

	Describe("Extract()", func() {
		Context("With a valid stemcell", func() {
			It("returns a file handle to the VMDK contained in a vSphere stemcell", func() {


			})
		})
		Context("When the stemcell is invalid", func() {
			It("returns an error", func() {

			})
		})
		Context("When the stemcell does not exist", func() {
			It("returns an error", func() {

			})
		})

	})
	Describe("Cleanup()", func() {
		It("cleans up the temp directory and closes the file handle?", func() {

		})
	})
})
