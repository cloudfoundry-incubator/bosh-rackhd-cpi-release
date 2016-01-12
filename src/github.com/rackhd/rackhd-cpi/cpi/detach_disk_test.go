package cpi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	. "github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/helpers"
)

var _ = Describe("DetachDisk", func() {

	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":1}`, serverURL.Host))
		cpiConfig, err = config.New(jsonReader)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("given a disk CID that exists", func() {
		Context("given a disk CID that is not attached", func() {
			It("returns an error", func() {
				jsonInput := []byte(`[
						"valid_vm_cid_1",
						"valid_disk_cid_1"
					]`)
				var extInput bosh.MethodArguments
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).ToNot(HaveOccurred())

				expectedNodes := helpers.LoadNodes("../spec_assets/dummy_attached_disk_response.json")
				expectedNodesData, err := json.Marshal(expectedNodes)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/common/nodes"),
						ghttp.RespondWith(http.StatusOK, expectedNodesData),
					),
				)

				err = DetachDisk(cpiConfig, extInput)
				Expect(err).To(MatchError("Disk: valid_disk_cid_1 is detached\n"))
				Expect(len(server.ReceivedRequests())).To(Equal(1))
			})
		})

		Context("given a disk that is attached", func() {
			Context("when given a vm cid that the disk does not belong to", func() {
				It("returns an error", func() {
					jsonInput := []byte(`[
							"invalid_vm_cid_2",
							"valid_disk_cid_2"
						]`)
					var extInput bosh.MethodArguments
					err := json.Unmarshal(jsonInput, &extInput)
					Expect(err).ToNot(HaveOccurred())

					expectedNodes := helpers.LoadNodes("../spec_assets/dummy_attached_disk_response.json")
					expectedNodesData, err := json.Marshal(expectedNodes)
					Expect(err).ToNot(HaveOccurred())
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/common/nodes"),
							ghttp.RespondWith(http.StatusOK, expectedNodesData),
						),
					)

					err = DetachDisk(cpiConfig, extInput)
					Expect(err).To(MatchError("Disk valid_disk_cid_2 does not belong to VM invalid_vm_cid_2\n"))
					Expect(len(server.ReceivedRequests())).To(Equal(1))
				})
			})

			Context("given a VM CID that the disk belongs to", func() {
				It("detaches the disk", func() {
					jsonInput := []byte(`[
						"valid_vm_cid_2",
						"valid_disk_cid_2"
					]`)
					var extInput bosh.MethodArguments
					err := json.Unmarshal(jsonInput, &extInput)
					Expect(err).NotTo(HaveOccurred())

					expectedNodes := helpers.LoadNodes("../spec_assets/dummy_attached_disk_response.json")
					expectedNodesData, err := json.Marshal(expectedNodes)
					Expect(err).ToNot(HaveOccurred())

					body := rackhdapi.PersistentDiskSettingsContainer{
						PersistentDisk: rackhdapi.PersistentDiskSettings{
							DiskCID:    "valid_disk_cid_2",
							Location:   "/dev/sdb",
							IsAttached: false,
						},
					}
					bodyBytes, err := json.Marshal(body)

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/common/nodes"),
							ghttp.RespondWith(http.StatusOK, expectedNodesData),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79eb14e66816f6152fffb"),
							ghttp.VerifyJSON(string(bodyBytes)),
						),
					)

					err = DetachDisk(cpiConfig, extInput)
					Expect(len(server.ReceivedRequests())).To(Equal(2))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Context("given a nonexistent disk CID", func() {
		It("returns an error", func() {
			jsonInput := []byte(`[
					"valid_vm_cid_1",
					"invalid_disk_cid"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_attached_disk_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			err = DetachDisk(cpiConfig, extInput)
			Expect(err).To(MatchError("Disk: invalid_disk_cid not found\n"))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})
})
