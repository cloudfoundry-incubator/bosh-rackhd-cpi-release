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
	"github.com/rackhd/rackhd-cpi/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("DetachDisk", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.CREATE_DISK)
	})

	AfterEach(func() {
		server.Close()
	})

	Context("given a disk CID that exists", func() {
		Context("given a disk CID that is not attached", func() {
			It("returns an error", func() {
				vmCID := "vm_cid-fake_uuid"
				diskCID := "disk_cid-fake_uuid"
				jsonInput := []byte(`[
            "` + vmCID + `",
            "` + diskCID + `"
          ]`)
				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_detached.json")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
						ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
					),
				)

				err = cpi.DetachDisk(cpiConfig, extInput)
				errMsg := fmt.Sprintf("disk: %s is already detached to VM %s", diskCID, vmCID)
				Expect(err).To(MatchError(errMsg))
				Expect(len(server.ReceivedRequests())).To(Equal(1))
			})
		})

		Context("given a disk that is attached", func() {
			Context("when given a vm cid that the disk does not belong to", func() {
				It("returns an error", func() {
					vmCID := "vm_cid-fake_uuid"
					diskCID := "disk_cid-not_exist"
					jsonInput := []byte(`[
            	"` + vmCID + `",
            	"` + diskCID + `"
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

					err = cpi.DetachDisk(cpiConfig, extInput)
					errMsg := fmt.Sprintf("another disk is attached to VM %s", vmCID)
					Expect(err).To(MatchError(errMsg))
					Expect(len(server.ReceivedRequests())).To(Equal(1))
				})
			})

			Context("given a VM CID that the disk belongs to", func() {
				It("detaches the disk", func() {
					vmCID := "vm_cid-fake_uuid"
					diskCID := "disk_cid-fake_uuid"
					jsonInput := []byte(`[
            	"` + vmCID + `",
            	"` + diskCID + `"
          	]`)
					var extInput bosh.MethodArguments
					err := json.Unmarshal(jsonInput, &extInput)
					Expect(err).ToNot(HaveOccurred())

					expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_attached.json")
					var expectedNodes []models.TagNode
					err = json.Unmarshal(expectedNodesBytes, &expectedNodes)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(expectedNodes)).To(Equal(1))

					node := expectedNodes[0]
					node.PersistentDisk.IsAttached = false
					expectedPersistanceSettingsBytes, err := json.Marshal(map[string]interface{}{
						"persistent_disk": node.PersistentDisk,
					})

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
							ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/2.0/nodes/%s", node.ID)),
							ghttp.VerifyJSON(string(expectedPersistanceSettingsBytes)),
							ghttp.RespondWith(http.StatusOK, nil),
						),
					)

					err = cpi.DetachDisk(cpiConfig, extInput)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(server.ReceivedRequests())).To(Equal(2))
				})
			})
		})
	})

	Context("given a nonexistent disk CID", func() {
		It("returns an error", func() {
			vmCID := "vm_cid-fake_uuid"
			diskCID := "disk_cid-not_exist"
			jsonInput := []byte(`[
					"` + vmCID + `",
          "` + diskCID + `"
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

			err = cpi.DetachDisk(cpiConfig, extInput)
			errMsg := fmt.Sprintf("disk: %s was not found on VM %s", diskCID, vmCID)
			Expect(err).To(MatchError(errMsg))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})
})
