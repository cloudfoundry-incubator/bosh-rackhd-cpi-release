package cpi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
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
		var expectedNodes []models.Node

		BeforeEach(func() {
			expectedNodes = helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/2.0/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)
		})

		Context("when there is a persistent disk left before deprovisioning", func() {
			It("deprovisions the node", func() {
				jsonInput := []byte(`["vm-5678"]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					helpers.MakeWorkflowHandlers(
						"Deprovision",
						cpiConfig.RequestID,
						"55e79ea54e66816f6152fff9",
					)...,
				)

				err = cpi.DeleteVM(cpiConfig, extInput)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(8))
			})
		})

		Context("when there is no persistent disk left before deprovisioning", func() {
			It("deprovisions the node and sets the status to available", func() {
				jsonInput := []byte(`["vm-1234"]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					helpers.MakeWorkflowHandlers(
						"Deprovision",
						cpiConfig.RequestID,
						"55e79eb14e66816f6152fffb",
					)...,
				)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/2.0/nodes/55e79eb14e66816f6152fffb"),
						ghttp.VerifyJSON("{\"status\": \"available\"}"),
					),
				)

				err = cpi.DeleteVM(cpiConfig, extInput)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(9))
			})
		})

		Context("when there are attached disks to a VM", func() {
			It("detaches a disk from the VM and deletes the VM", func() {
				jsonInput := []byte(`["valid_vm_cid_2"]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())
				nodeID := "5665a65a0561790005b77b85"
				container := rackhdapi.PersistentDiskSettingsContainer{
					PersistentDisk: rackhdapi.PersistentDiskSettings{
						DiskCID:    fmt.Sprintf("%s-%s", nodeID, cpiConfig.RequestID),
						Location:   fmt.Sprintf("/dev/%s", rackhdapi.PersistentDiskLocation),
						IsAttached: false,
					},
				}
				expectedPersistentDiskSettings, err := json.Marshal(container)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
						ghttp.VerifyJSON(string(expectedPersistentDiskSettings)),
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
