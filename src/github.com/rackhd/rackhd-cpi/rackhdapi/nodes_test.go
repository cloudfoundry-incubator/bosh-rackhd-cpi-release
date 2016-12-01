package rackhdapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Nodes", func() {
	var server *ghttp.Server
	var c config.Cpi

	BeforeEach(func() {
		server, _, c, _ = helpers.SetUp("")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleaseNode", func() {
		It("return a node with reserved flag unset", func() {
			c := config.Cpi{ApiServer: os.Getenv("RACKHD_API_URL")}
			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())
			targetNodeID := nodes[0].ID

			err = rackhdapi.CreateTag(c, targetNodeID, "unavailable")
			Expect(err).ToNot(HaveOccurred())
			tags, err := rackhdapi.GetTags(c, targetNodeID)
			Expect(tags).To(Equal([]string{"unavailable"}))


			err = rackhdapi.ReleaseNode(c, targetNodeID)
			Expect(err).ToNot(HaveOccurred())

			tags, err = rackhdapi.GetTags(c, targetNodeID)
			Expect(err).ToNot(HaveOccurred())
			Expect(tags).ToNot(ContainElement(models.Unavailable))
		})
	})

	Describe("Getting nodes", func() {
		It("return expected nodes' fields", func() {
			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_two_node_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/2.0/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(nodes).To(Equal(expectedNodes))
		})
	})

	Describe("Getting a single node by nodeID", func() {
		XIt("returns node with the nodeID specified", func() {
			expectedNode := helpers.LoadNode("../spec_assets/dummy_create_vm_with_disk_response.json")
			expectedNodeData, err := json.Marshal(expectedNode)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/2.0/nodes/5665a65a0561790005b77b85"),
					ghttp.RespondWith(http.StatusOK, expectedNodeData),
				),
			)

			node, err := rackhdapi.GetNode(c, "5665a65a0561790005b77b85")
			Expect(err).ToNot(HaveOccurred())
			Expect(node).To(Equal(expectedNode))
		})
	})

	Describe("GetOBMSettings", func() {
		It("returns a node's OBM settings", func() {
			dummyResponsePath := "../spec_assets/dummy_one_node_response.json"
			httpResponse := helpers.LoadJSON(dummyResponsePath)
			expectedResponse := helpers.LoadNode(dummyResponsePath)

			nodeID := "nodeID"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
					ghttp.RespondWith(http.StatusOK, httpResponse),
				),
			)

			response, err := rackhdapi.GetOBMSettings(c, nodeID)
			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(response).To(Equal(expectedResponse.OBMS))
		})
	})

	Describe("Getting catalog", func() {
		It("returns a catalog", func() {
			expectedNodeCatalog := helpers.LoadNodeCatalog("../spec_assets/dummy_node_catalog_response.json")
			expectedNodeCatalogData, err := json.Marshal(expectedNodeCatalog)
			testNodeID := "55e79eb14e66816f6152fffb"
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s/catalogs/ohai", testNodeID)),
					ghttp.RespondWith(http.StatusOK, expectedNodeCatalogData),
				),
			)

			catalog, err := rackhdapi.GetNodeCatalog(c, testNodeID)
			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(catalog).To(Equal(expectedNodeCatalog))
		})
	})

	Describe("Setting node metadata", func() {
		XIt("Adds metadata to the node", func() {
			nodeID := "node_id"
			metadata := "{}"

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
					ghttp.VerifyJSON(fmt.Sprintf(`{"metadata": %s}`, metadata)),
				),
			)

			err := rackhdapi.SetNodeMetadata(c, nodeID, metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
