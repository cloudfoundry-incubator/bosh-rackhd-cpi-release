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
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

var _ = Describe("DeleteVM", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_reserve_node_attempts":1, "request_id": "my_request_id"}`, serverURL.Host))
		request = bosh.CpiRequest{Method: bosh.DELETE_VM}
		cpiConfig, err = config.New(jsonReader, request)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("with a valid VM CID and valid states", func() {
		var extInput bosh.MethodArguments
		var expectedNodes []rackhdapi.Node

		BeforeEach(func() {
			expectedNodes = helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
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
						"my_request_id",
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
						"my_request_id",
						"55e79eb14e66816f6152fffb",
					)...,
				)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79eb14e66816f6152fffb"),
						ghttp.VerifyJSON("{\"status\": \"available\"}"),
					),
				)

				err = cpi.DeleteVM(cpiConfig, extInput)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(9))
			})
		})
	})
})
