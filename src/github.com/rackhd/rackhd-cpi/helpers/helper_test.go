package helpers_test

import (
  "github.com/rackhd/rackhd-cpi/helpers"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
)

var _ = Describe("Helper", func() {
  Describe("Transform", func() {
    Context("when the array is not empty", func() {
      It("should convert to bytes", func() {
        a := []string{"apple", "pinapple", "pen"}
        b := helpers.ArrayToBytes(a)
        Expect(b).To(Equal([]byte("[\"apple\",\"pinapple\",\"pen\"]")))
      })
    })
  })
})
