package rackhdapi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Nodes", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi

	BeforeEach(func() {
		server, jsonReader, cpiConfig, _ = helpers.SetUp("")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleaseNode", func() {
		It("return a node with reserved flag unset", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			c := config.Cpi{ApiServer: apiServer}

			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())
			targetNodeID := nodes[0].ID
			err = rackhdapi.ReleaseNode(c, targetNodeID)
			Expect(err).ToNot(HaveOccurred())
			nodeURL := fmt.Sprintf("%s/api/common/nodes/%s", c.ApiServer, targetNodeID)

			resp, err := http.Get(nodeURL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			nodeBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var node rackhdapi.Node
			err = json.Unmarshal(nodeBytes, &node)
			Expect(err).ToNot(HaveOccurred())
			Expect(node.Status).To(Equal(rackhdapi.Available))
		})
	})

	Describe("Getting nodes", func() {
		It("return expected nodes' fields", func() {
			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_two_node_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			nodes, err := rackhdapi.GetNodes(cpiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(nodes).To(Equal(expectedNodes))
		})
	})

	Describe("Getting a single node by CID", func() {
		It("returns expected node's fields", func() {
			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_all_nodes_are_vms.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			node, err := rackhdapi.GetNodeByVMCID(cpiConfig, "vm-5678")

			Expect(err).ToNot(HaveOccurred())
			Expect(node).To(Equal(expectedNodes[0]))
		})
	})

	Describe("Getting a single node by nodeID", func() {
		It("returns node with the nodeID specified", func() {
			expectedNode := helpers.LoadNode("../spec_assets/dummy_create_vm_with_disk_response.json")
			expectedNodeData, err := json.Marshal(expectedNode)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes/5665a65a0561790005b77b85"),
					ghttp.RespondWith(http.StatusOK, expectedNodeData),
				),
			)

			node, err := rackhdapi.GetNode(cpiConfig, "5665a65a0561790005b77b85")

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
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodeID)),
					ghttp.RespondWith(http.StatusOK, httpResponse),
				),
			)

			response, err := rackhdapi.GetOBMSettings(cpiConfig, nodeID)
			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(response).To(Equal(expectedResponse.OBMSettings))
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
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s/catalogs/ohai", testNodeID)),
					ghttp.RespondWith(http.StatusOK, expectedNodeCatalogData),
				),
			)

			catalog, err := rackhdapi.GetNodeCatalog(cpiConfig, testNodeID)

			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(catalog).To(Equal(expectedNodeCatalog))
		})
	})

	Describe("blocking nodes", func() {
		It("sends a request to block a node", func() {
			nodes := helpers.LoadNodes("../spec_assets/dummy_two_node_response.json")

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/common/nodes/%s", nodes[0].ID)),
					ghttp.VerifyJSON(fmt.Sprintf(`{"status": "%s", "status_reason": "%s"}`, "blocked", "Node has missing disks")),
				),
			)

			rackhdapi.BlockNode(cpiConfig, nodes[0].ID)
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Describe("Setting node metadata", func() {
		It("Adds metadata to the node", func() {
			nodeID := "node_id"
			metadata := "{}"

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/common/nodes/%s", nodeID)),
					ghttp.VerifyJSON(fmt.Sprintf(`{"metadata": %s}`, metadata)),
				),
			)

			err := rackhdapi.SetNodeMetadata(cpiConfig, nodeID, metadata)
			Expect(err).ToNot(HaveOccurred())

			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
