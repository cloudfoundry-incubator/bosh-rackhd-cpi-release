package cpi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteVM", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.DELETE_VM)
		cpiConfig.RequestID = "requestid"
	})

	AfterEach(func() {
		server.Close()
	})

	Context("with a valid VM CID and valid states", func() {
		var extInput bosh.MethodArguments

		Context("when there is a persistent disk left before deprovisioning", func() {
			It("deprovisions the node", func() {
				vmCID := "vm_cid-fake_uuid"
				expectedNodesData := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_detached.json")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
						ghttp.RespondWith(http.StatusOK, expectedNodesData),
					),
				)

				jsonInput := []byte(`["` + vmCID + `"]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					helpers.MakeWorkflowHandlers(
						"Deprovision",
						cpiConfig.RequestID,
						"57fb9fb03fcc55c807add41c",
					)...,
				)

				err = cpi.DeleteVM(cpiConfig, extInput)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(8))
			})
		})

		Context("when there is no persistent disk left before deprovisioning", func() {
			It("deprovisions the node and sets the status to available", func() {
				expectedNodesData := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_cid.json")
				vmCID := "vm-1234"
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
						ghttp.RespondWith(http.StatusOK, expectedNodesData),
					),
				)

				jsonInput := []byte(`["` + vmCID + `"]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					helpers.MakeWorkflowHandlers(
						"Deprovision",
						cpiConfig.RequestID,
						"57fb9fb03fcc55c807add41c",
					)...,
				)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/2.0/nodes/57fb9fb03fcc55c807add41c/tags/"+models.Unavailable),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/2.0/nodes/57fb9fb03fcc55c807add41c/tags/"+vmCID),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/api/2.0/nodes/57fb9fb03fcc55c807add41c/tags/57fb9fb03fcc55c807add41c"),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err = cpi.DeleteVM(cpiConfig, extInput)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(9))
			})
		})

		Context("when there are attached disks to a VM", func() {
			It("detaches a disk from the VM and deletes the VM", func() {
				vmCID := "vm_cid-fake_uuid"

				jsonInput := []byte(`["` + vmCID + `"]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				expectedNodesData := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_attached.json")
				var expectedNodes []models.TagNode
				err = json.Unmarshal(expectedNodesData, &expectedNodes)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(expectedNodes)).To(Equal(1))

				expectedNode := expectedNodes[0]
				expectedNode.PersistentDisk.IsAttached = false
				nodeID := expectedNode.ID

				expectedPersistentDiskSettings, err := json.Marshal(map[string]interface{}{
					"persistent_disk": expectedNode.PersistentDisk,
				})
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", vmCID)),
						ghttp.RespondWith(http.StatusOK, expectedNodesData),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
						ghttp.VerifyJSON(string(expectedPersistentDiskSettings)),
						ghttp.RespondWith(http.StatusOK, nil),
					),
				)
				server.AppendHandlers(
					helpers.MakeWorkflowHandlers(
						"Deprovision",
						cpiConfig.RequestID,
						nodeID,
					)...,
				)
				err = cpi.DeleteVM(cpiConfig, extInput)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(9))
			})
		})
	})
})
