package rackhdapi_test

import (
	"fmt"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tags", func() {
	var cpiConfig config.Cpi

	BeforeEach(func ()  {
		apiServer, err := helpers.GetRackHDHost()
		Expect(err).ToNot(HaveOccurred())
		cpiConfig = config.Cpi{ApiServer: apiServer}
	})

	Describe("GetTags", func ()  {
		Context("when there are tags attached to nodes", func ()  {
			FIt("should return all tags without error", func ()  {
				nodeID := "57fb9fb03fcc55c807add41c"
				tags, err := rackhdapi.GetTags(cpiConfig, nodeID)
				Expect(err).ToNot(HaveOccurred())
				fmt.Printf("\ntags: %+v\n", tags)
			})
		})
	})

	// Describe("DeleteTag", func ()  {
	// 	Context("when deleting an non-exsiting tag", func ()  {
	// 		FIt("should return with error", func ()  {
	// 			nodeID := "57fb9fb03fcc55c807add41c"
	// 			tag := "fake-tag"
	// 			err := rackhdapi.DeleteTag(cpiConfig, nodeID, tag)
	// 			Expect(err).To(HaveOccurred())
	// 		})
	// 	})
	// })
	//
	// Describe("UpdateTag", func ()  {
	// 	Context("when create a new tag", func ()  {
	// 		It("should return without error", func ()  {
	// 			nodeID := "57fb9fb03fcc55c807add41c"
	// 			err := rackhdapi.UpdateTag(cpiConfig, nodeID, "fake-tag")
	// 			Expect(err).ToNot(HaveOccurred())
	// 		})
	// 	})
	// })
})
