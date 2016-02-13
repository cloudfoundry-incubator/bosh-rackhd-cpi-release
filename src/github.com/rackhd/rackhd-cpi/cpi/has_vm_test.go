package cpi_test

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/cpi"
	"github.com/rackhd/rackhd-cpi/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Cpi/HasVm", func() {
	Context("Has VM", func() {
		var server *ghttp.Server
		var jsonReader *strings.Reader
		var cpiConfig config.Cpi
		var request bosh.CpiRequest

		BeforeEach(func() {
			server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.HAS_VM)
		})

		AfterEach(func() {
			server.Close()
		})

		It("Find a vm that exist", func() {
			cid := "vm-1234"

			var metadataInput bosh.MethodArguments
			metadataInput = append(metadataInput, cid)
			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_two_node_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			hasVM, err := cpi.HasVM(cpiConfig, metadataInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(hasVM).To(BeTrue())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("Cannot find a vm that does not exist", func() {
			cid := "does-not-exist-cid"

			var metadataInput bosh.MethodArguments
			metadataInput = append(metadataInput, cid)
			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_two_node_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			hasVM, err := cpi.HasVM(cpiConfig, metadataInput)
			Expect(err).ToNot(HaveOccurred())
			Expect(hasVM).To(BeFalse())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
