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
		It("returns true", func() {
			jsonInput := []byte(`[
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

			result, err := HasDisk(cpiConfig, extInput)
			Expect(err).To(BeNil())
			Expect(result).To(BeTrue())
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})

	Context("given a disk CID that not exists", func() {
		It("returns false", func() {
			jsonInput := []byte(`[
					"invalid_disk_cid_1"
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

			result, err := HasDisk(cpiConfig, extInput)
			Expect(err).To(BeNil())
			Expect(result).To(BeFalse())
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})

	Context("given a disk CID that is an empty string", func() {
		It("returns false", func() {
			jsonInput := []byte(`[
					""
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			result, err := HasDisk(cpiConfig, extInput)
			Expect(err).To(BeNil())
			Expect(result).To(BeFalse())
		})
	})
})
