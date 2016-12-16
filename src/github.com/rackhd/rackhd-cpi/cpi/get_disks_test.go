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

var _ = Describe("GetDisks", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.GET_DISKS)
	})

	AfterEach(func() {
		server.Close()
	})

	Context("given a vm CID that exists", func() {
		Context("the vm has persistent disk", func() {
			It("returns disk CID", func() {
				vmCID := "vm_cid-fake_uuid"
				jsonInput := []byte(`[
            "` + vmCID + `"
          ]`)
				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_attached.json")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
						ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
					),
				)

				result, err := cpi.GetDisks(cpiConfig, extInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal([]string{"disk_cid-fake_uuid"}))
				Expect(len(server.ReceivedRequests())).To(Equal(1))
			})
		})

		Context("the vm does not have persistent disk", func() {
			It("returns empty array", func() {
				vmCID := "vm_cid-fake_uuid"
				jsonInput := []byte(`[
            "` + vmCID + `"
          ]`)
				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_cid.json")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
						ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
					),
				)

				result, err := cpi.GetDisks(cpiConfig, extInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal([]string{}))
				Expect(len(server.ReceivedRequests())).To(Equal(1))
			})
		})
	})

	Context("given a vm CID that not exists", func() {
		It("returns an error", func() {
			vmCID := "vm_cid-not_exist"
			jsonInput := []byte(`[
          "` + vmCID + `"
        ]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
			)

			_, err = cpi.GetDisks(cpiConfig, extInput)
			Expect(err).To(HaveOccurred())
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})
})
