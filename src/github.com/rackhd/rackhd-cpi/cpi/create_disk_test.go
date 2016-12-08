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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateDisk", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.CREATE_DISK)
		cpiConfig.RequestID = "my_id"
	})

	AfterEach(func() {
		server.Close()
	})

	Context("If VM cid cannot be found", func() {
		It("returns error", func() {
			jsonInput := []byte(`[
					25000,
					{
						"some": "options"
					},
					"invalid-vm-cid"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).NotTo(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/2.0/tags/invalid-vm-cid/nodes"),
					ghttp.RespondWith(http.StatusOK, []byte("[]")),
				),
			)

			diskCID, err := cpi.CreateDisk(cpiConfig, extInput)
			Expect(err).To(HaveOccurred())
			Expect(diskCID).To(Equal(""))
		})
	})

	Context("If there is already a disk on the VM", func() {
		It("returns error", func() {
			jsonInput := []byte(`[
					2500,
					{
						"some": "options"
					},
					"vm-5678"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).NotTo(HaveOccurred())

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/2.0/tags/vm-5678/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			diskCID, err := cpi.CreateDisk(cpiConfig, extInput)
			Expect(err).To(HaveOccurred())
			Expect(diskCID).To(Equal(""))
		})
	})

	Context("there is an available disk", func() {
		Context("without enough disk space", func() {
			It("returns an error", func() {
				jsonInput := []byte(`[
							250000000000,
							{
								"some": "options"
							},
							"vm-1234"
						]`)
				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				expectedNodes := helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
				expectedNodesData, err := json.Marshal(expectedNodes)
				Expect(err).ToNot(HaveOccurred())
				expectedNodeCatalog := helpers.LoadNodeCatalog("../spec_assets/dummy_node_catalog_response.json")
				expectedNodeCatalogData, err := json.Marshal(expectedNodeCatalog)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/2.0/tags/vm-1234/nodes"),
						ghttp.RespondWith(http.StatusOK, expectedNodesData),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/2.0/nodes/55e79eb14e66816f6152fffb/catalogs/ohai"),
						ghttp.RespondWith(http.StatusOK, expectedNodeCatalogData),
					),
				)

				diskCID, err := cpi.CreateDisk(cpiConfig, extInput)
				Expect(err).To(HaveOccurred())
				Expect(diskCID).To(Equal(""))
			})
		})

		Context("with enough disk space", func() {
			Context("If VM cid is empty", func() {
				It("creates the disk and returns the disk cid", func() {
					jsonInput := []byte(`[
								2500,
								{
									"some": "options"
								},
								""
							]`)
					var extInput bosh.MethodArguments
					err := json.Unmarshal(jsonInput, &extInput)
					Expect(err).NotTo(HaveOccurred())

					expectedPersistentDiskSettings := `{
             "persistent_disk": {
               "pregenerated_disk_cid": "",
               "disk_cid": "57fb9fb03fcc55c807add402-my_id",
               "location": "/dev/sdb",
               "attached": false
             }
           }`

					server.AppendHandlers(
						helpers.MakeTryReservationHandlers(
							"my_id",
							"57fb9fb03fcc55c807add402",
							"../spec_assets/dummy_create_disk_nodes_response.json",
							"../spec_assets/dummy_create_disk_catalog_response.json",
						)...,
					)
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PATCH", "/api/2.0/nodes/57fb9fb03fcc55c807add402"),
							ghttp.VerifyJSON(expectedPersistentDiskSettings),
						),
					)

					diskCID, err := cpi.CreateDisk(cpiConfig, extInput)
					Expect(err).ToNot(HaveOccurred())
					Expect(diskCID).ToNot(Equal(""))
				})
			})

			Context("If VM CID is not empty", func() {
				It("creates the disk and returns the disk cid", func() {
					jsonInput := []byte(`[
								2500,
								{
									"some": "options"
								},
								"vm-1234"
							]`)
					var extInput bosh.MethodArguments
					err := json.Unmarshal(jsonInput, &extInput)
					Expect(err).NotTo(HaveOccurred())

					expectedNodes := helpers.LoadTagNodes("../spec_assets/tag_node_with_cid.json")
					expectedNodesBytes, err := json.Marshal(expectedNodes)
					Expect(err).ToNot(HaveOccurred())
					expectedNodeCatalogBytes := helpers.LoadJSON("../spec_assets/dummy_node_catalog_response.json")

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/2.0/tags/vm-1234/nodes"),
							ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s/catalogs/ohai", expectedNodes[0].ID)),
							ghttp.RespondWith(http.StatusOK, expectedNodeCatalogBytes),
						),
						ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/2.0/nodes/%s", expectedNodes[0].ID)),
					)

					diskCID, err := cpi.CreateDisk(cpiConfig, extInput)
					Expect(err).ToNot(HaveOccurred())
					Expect(diskCID).ToNot(Equal(""))
				})
			})
		})
	})
})
