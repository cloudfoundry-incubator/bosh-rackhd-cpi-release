package cpi_test

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	. "github.com/rackhd/rackhd-cpi/cpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/helpers"
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
				jsonInput := []byte(`[
						"valid_vm_cid_1"
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

				result, err := GetDisks(cpiConfig, extInput)
				Expect(err).To(BeNil())
				expectedResult := []string{"valid_disk_cid_1"}
				Expect(result).To(Equal(expectedResult))
				Expect(len(server.ReceivedRequests())).To(Equal(1))
			})
		})

		Context("the vm does not have persistent disk", func() {
			It("returns empty array", func() {
				jsonInput := []byte(`[
						"valid_vm_cid_3"
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

				result, err := GetDisks(cpiConfig, extInput)
				Expect(err).To(BeNil())
				expectedResult := make([]string, 0)
				Expect(result).To(Equal(expectedResult))
				Expect(len(server.ReceivedRequests())).To(Equal(1))
			})
		})
	})

	Context("given a vm CID that not exists", func() {
		It("returns an error", func() {
			jsonInput := []byte(`[
					"invalid_vm_cid_3"
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

			_, err = GetDisks(cpiConfig, extInput)
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError("vm with cid: invalid_vm_cid_3 was not found"))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})
})
