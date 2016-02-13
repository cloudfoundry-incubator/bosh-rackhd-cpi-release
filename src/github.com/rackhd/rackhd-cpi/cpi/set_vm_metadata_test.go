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
)

var _ = Describe("Setting VM Metadata", func() {
	Context("When called with metadata", func() {
		var server *ghttp.Server
		var jsonReader *strings.Reader
		var cpiConfig config.Cpi
		var request bosh.CpiRequest

		BeforeEach(func() {
			server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.SET_VM_METADATA)
		})

		AfterEach(func() {
			server.Close()
		})

		It("Sends a request to set metadata to the RackHD API", func() {

			id := "55e79ea54e66816f6152fff9"
			cid := "vm-5678"
			metadata := map[string]interface{}{
				"stuff":  "definitely",
				"thing1": 3563456,
				"thing2": "bloop",
			}

			var metadataInput bosh.MethodArguments
			metadataInput = append(metadataInput, cid)
			metadataInput = append(metadataInput, metadata)

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
				ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/common/nodes/%s", id)),
			)

			err = cpi.SetVMMetadata(cpiConfig, metadataInput)
			Expect(err).ToNot(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(2))
		})
	})
})
