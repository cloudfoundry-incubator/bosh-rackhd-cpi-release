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
		Context("given a disk CID for an already attached disk", func() {
			Context("given a disk CID that is same as attached disk", func() {
				It("returns an error", func() {
					jsonInput := []byte(`[
              "vm_cid-fake_uuid",
              "disk_cid-fake_uuid"
            ]`)
					var extInput bosh.MethodArguments
					err := json.Unmarshal(jsonInput, &extInput)
					Expect(err).ToNot(HaveOccurred())

					vmCID := "vm_cid-fake_uuid"
					expectedNodesData := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_attached.json")
					Expect(err).ToNot(HaveOccurred())
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
							ghttp.RespondWith(http.StatusOK, expectedNodesData),
						),
					)

					err = cpi.AttachDisk(cpiConfig, extInput)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(server.ReceivedRequests())).To(Equal(1))
				})
			})

			Context("given a disk CID that is not same as attached disk", func() {
				Context("if existing disk is attached", func() {
					It("returns an error", func() {
						vmCID := "vm_cid-fake_uuid"
						attachedDiskCID := "disk_cid-fake_uuid"
						unattachedDiskCID := "disk_cid-fake_uuid_unattached"

						jsonInput := []byte(fmt.Sprintf(`[
                "%s",
                "%s"
              ]`, vmCID, unattachedDiskCID))
						var extInput bosh.MethodArguments
						err := json.Unmarshal(jsonInput, &extInput)
						Expect(err).ToNot(HaveOccurred())

						expectedNodesData := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_attached.json")
						Expect(err).ToNot(HaveOccurred())
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
								ghttp.RespondWith(http.StatusOK, expectedNodesData),
							),
						)

						err = cpi.AttachDisk(cpiConfig, extInput)
						errMsg := fmt.Sprintf(
							"node %s has persistent disk %s attached. Cannot attach additional disk %s",
							vmCID, attachedDiskCID, unattachedDiskCID)
						Expect(err).To(MatchError(errMsg))
						Expect(len(server.ReceivedRequests())).To(Equal(1))
					})
				})

				Context("if existing disk is NOT attached", func() {
					It("returns an error", func() {
						vmCID := "valid_vm_cid_5"
						jsonInput := []byte(fmt.Sprintf(`[
                "%s",
                "new_disk_cid"
              ]`, vmCID))
						var extInput bosh.MethodArguments
						err := json.Unmarshal(jsonInput, &extInput)
						Expect(err).ToNot(HaveOccurred())

						expectedNodesData := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_detached.json")
						Expect(err).ToNot(HaveOccurred())
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
								ghttp.RespondWith(http.StatusOK, expectedNodesData),
							),
						)

						err = cpi.AttachDisk(cpiConfig, extInput)
						Expect(err).To(MatchError("node valid_vm_cid_5 has persistent disk disk_cid-fake_uuid, but detached. Cannot attach disk new_disk_cid"))
						Expect(len(server.ReceivedRequests())).To(Equal(1))
					})
				})
			})
		})

		Context("given a disk that is not attached", func() {
			Context("given a VM CID that the disk belongs to", func() {
				It("attaches the disk", func() {
					vmCID := "vm_cid-fake_uuid"
					diskCID := "disk_cid-fake_uuid"

					jsonInput := []byte(fmt.Sprintf(`[
            "%s",
            "%s"
          ]`, vmCID, diskCID))
					var extInput bosh.MethodArguments
					err := json.Unmarshal(jsonInput, &extInput)
					Expect(err).NotTo(HaveOccurred())

					expectedNodeBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_detached.json")
					var expectedNodes []models.TagNode
					err = json.Unmarshal(expectedNodeBytes, &expectedNodes)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(expectedNodes)).To(Equal(1))
					expectedNode := expectedNodes[0]
					expectedNode.PersistentDisk.IsAttached = true

					bodyBytes, err := json.Marshal(map[string]interface{}{
						"persistent_disk": expectedNode.PersistentDisk,
					})
					Expect(err).NotTo(HaveOccurred())

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
							ghttp.RespondWith(http.StatusOK, expectedNodeBytes),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/2.0/nodes/%s", expectedNode.ID)),
							ghttp.VerifyJSON(string(bodyBytes)),
						),
					)

					err = cpi.AttachDisk(cpiConfig, extInput)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(server.ReceivedRequests())).To(Equal(2))
				})
			})
		})
	})

	Context("given a nonexistent disk CID", func() {
		It("returns an error", func() {
			vmCID := "valid_vm_cid_3"
			jsonInput := []byte(fmt.Sprintf(`[
          "%s",
          "invalid_disk_cid"
        ]`, vmCID))
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodesData := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_cid.json")
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			err = cpi.AttachDisk(cpiConfig, extInput)
			Expect(err).To(MatchError("disk: invalid_disk_cid not found on VM: valid_vm_cid_3"))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})
})
