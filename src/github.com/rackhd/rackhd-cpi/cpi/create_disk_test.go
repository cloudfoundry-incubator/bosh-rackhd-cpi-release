package cpi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"
)

var _ = Describe("CreateDisk", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_reserve_node_attempts":1, "request_id": "my_id"}`, serverURL.Host))
		request = bosh.CpiRequest{Method: bosh.CREATE_DISK}
		cpiConfig, err = config.New(jsonReader, request)
		Expect(err).ToNot(HaveOccurred())
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

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_two_node_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
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
					25000,
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
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
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
						ghttp.VerifyRequest("GET", "/api/common/nodes"),
						ghttp.RespondWith(http.StatusOK, expectedNodesData),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/common/nodes/55e79eb14e66816f6152fffb/catalogs/ohai"),
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
								25000,
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
							"disk_cid": "55e79ea54e66816f6152fff9",
							"location": "/dev/sdb",
							"attached": false
						}
					}`

					server.AppendHandlers(
						helpers.MakeTryReservationHandlers(
							"my_id",
							"55e79ea54e66816f6152fff9",
							"../spec_assets/dummy_create_disk_nodes_response.json",
							"../spec_assets/dummy_create_disk_catalog_response.json",
						)...,
					)
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79ea54e66816f6152fff9"),
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
								25000,
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
							ghttp.VerifyRequest("GET", "/api/common/nodes"),
							ghttp.RespondWith(http.StatusOK, expectedNodesData),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/common/nodes/55e79eb14e66816f6152fffb/catalogs/ohai"),
							ghttp.RespondWith(http.StatusOK, expectedNodeCatalogData),
						),
						ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79eb14e66816f6152fffb"),
					)

					diskCID, err := cpi.CreateDisk(cpiConfig, extInput)
					Expect(err).ToNot(HaveOccurred())
					Expect(diskCID).ToNot(Equal(""))
				})
			})
		})
	})
})
