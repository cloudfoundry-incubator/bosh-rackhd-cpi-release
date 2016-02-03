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

var _ = Describe("DeleteVM", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi
	var request bosh.CpiRequest

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":1}`, serverURL.Host))
		request = bosh.CpiRequest{Method: bosh.DELETE_VM}
		cpiConfig, err = config.New(jsonReader, request)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("with a valid VM CID and valid states", func() {
		jsonInput := []byte(`["vm-5678"]`)
		taskStubData := []byte(`[{"injectableName": "Task.BOSH.DeprovisionNode.vm-5678"}]`)
		workflowStubData := []byte(`[{"injectableName": "Graph.BOSH.DeprovisionNode.vm-5678"}]`)
		nodeStubData := []byte(`{"obmSettings": [{"service": "fake-obm-service"}]}`)
		completedWorkflowResponse := []byte(`{"id": "my_id", "_status": "succeeded"}`)

		It("deletes a VM", func() {
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
				ghttp.VerifyRequest("PUT", "/api/1.1/workflows/tasks"),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/1.1/workflows/tasks/library"),
					ghttp.RespondWith(http.StatusOK, taskStubData),
				),
				ghttp.VerifyRequest("PUT", "/api/1.1/workflows"),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/1.1/workflows/library"),
					ghttp.RespondWith(http.StatusOK, workflowStubData),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes/55e79ea54e66816f6152fff9"),
					ghttp.RespondWith(http.StatusOK, nodeStubData),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/1.1/nodes/55e79ea54e66816f6152fff9/workflows/"),
					ghttp.RespondWith(http.StatusCreated, completedWorkflowResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/workflows/my_id"),
					ghttp.RespondWith(http.StatusOK, completedWorkflowResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79ea54e66816f6152fff9"),
					ghttp.VerifyJSON("{\"status\": \"available\"}"),
				),
			)

			err = cpi.DeleteVM(cpiConfig, extInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(server.ReceivedRequests())).To(Equal(9))
		})
	})
})
