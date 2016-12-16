package cpi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("AttachDisk", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.ATTACH_DISK)
	})

	AfterEach(func() {
		server.Close()
	})

	Context("given a disk CID that exists", func() {
		It("returns true", func() {
			diskCID := "disk_cid-fake_uuid"
			jsonInput := []byte(`[
          "` + diskCID + `"
        ]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_disk_cid.json")
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", diskCID)),
					ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
				),
			)

			result, err := cpi.HasDisk(cpiConfig, extInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeTrue())
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})

	Context("given a disk CID that not exists", func() {
		It("returns false", func() {
			diskCID := "disk_cid-not_exist"
			jsonInput := []byte(`[
          "` + diskCID + `"
        ]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_disk_cid.json")
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", diskCID)),
					ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
				),
			)

			result, err := cpi.HasDisk(cpiConfig, extInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeFalse())
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})

	Context("given a disk CID that is an empty string", func() {
		It("returns false", func() {
			jsonInput := []byte(`[""]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			result, err := cpi.HasDisk(cpiConfig, extInput)
			Expect(err).To(BeNil())
			Expect(result).To(BeFalse())
		})
	})
})
