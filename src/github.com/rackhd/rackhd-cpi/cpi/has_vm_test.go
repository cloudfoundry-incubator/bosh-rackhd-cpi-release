package cpi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
			server = ghttp.NewServer()
			serverURL, err := url.Parse(server.URL())
			Expect(err).ToNot(HaveOccurred())
			jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":1}`, serverURL.Host))
			request = bosh.CpiRequest{Method: bosh.HAS_VM}
			cpiConfig, err = config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())
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
