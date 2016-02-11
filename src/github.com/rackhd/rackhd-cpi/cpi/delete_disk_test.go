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

var _ = Describe("DeleteDisk", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_reserve_node_attempts":1}`, serverURL.Host))
		request = bosh.CpiRequest{Method: bosh.DELETE_DISK}
		cpiConfig, err = config.New(jsonReader, request)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("when given a disk cid for an existing, unattached disk", func() {
		var extInput bosh.MethodArguments
		var expectedDeleteDiskBodyBytes []byte

		BeforeEach(func() {
			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_disks_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			container := rackhdapi.PersistentDiskSettingsContainer{
				PersistentDisk: rackhdapi.PersistentDiskSettings{},
			}
			expectedDeleteDiskBodyBytes, err = json.Marshal(container)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)
		})

		Context("when there is a VM left on the node", func() {
			It("deletes the disk", func() {
				jsonInput := []byte(`[
						"valid_disk_cid_3"
					]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79e9f4e66816f6152fff5"),
						ghttp.VerifyJSON(string(expectedDeleteDiskBodyBytes)),
					),
				)

				err = DeleteDisk(cpiConfig, extInput)
				Expect(len(server.ReceivedRequests())).To(Equal(2))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when there is no VM left on the node", func() {
			It("deletes the disk and sets the status to available", func() {
				jsonInput := []byte(`[
						"valid_disk_cid_1"
					]`)
				err := json.Unmarshal(jsonInput, &extInput)
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79ea54e66816f6152fff9"),
						ghttp.VerifyJSON(string(expectedDeleteDiskBodyBytes)),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79ea54e66816f6152fff9"),
						ghttp.VerifyJSON("{\"status\": \"available\"}"),
					),
				)

				err = DeleteDisk(cpiConfig, extInput)
				Expect(len(server.ReceivedRequests())).To(Equal(3))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("when given a disk cid for a non-existent disk", func() {
		It("returns an error", func() {
			jsonInput := []byte(`[
					"invalid_disk_cid"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_disks_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			err = DeleteDisk(cpiConfig, extInput)
			Expect(err).To(MatchError("Disk: invalid_disk_cid not found\n"))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})

	Context("when given a disk cid for a attached disk", func() {
		It("returns an error", func() {
			jsonInput := []byte(`[
					"valid_disk_cid_2"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_disks_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			err = DeleteDisk(cpiConfig, extInput)
			Expect(err).To(MatchError("Disk: valid_disk_cid_2 is attached\n"))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})
})
