package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func LoadJSON(nodePath string) []byte {
	dummyResponseFile, err := os.Open(nodePath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyResponseFile.Close()

	dummyResponseBytes, err := ioutil.ReadAll(dummyResponseFile)
	Expect(err).ToNot(HaveOccurred())

	return dummyResponseBytes
}

func LoadNodes(nodePath string) []rackhdapi.Node {
	dummyResponseBytes := LoadJSON(nodePath)

	nodes := []rackhdapi.Node{}
	err := json.Unmarshal(dummyResponseBytes, &nodes)
	Expect(err).ToNot(HaveOccurred())

	return nodes
}

func LoadNode(nodePath string) rackhdapi.Node {
	dummyResponseBytes := LoadJSON(nodePath)

	node := rackhdapi.Node{}
	err := json.Unmarshal(dummyResponseBytes, &node)
	Expect(err).ToNot(HaveOccurred())

	return node
}

func LoadNodeCatalog(nodeCatalogPath string) rackhdapi.NodeCatalog {
	dummyCatalogfile, err := os.Open(nodeCatalogPath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyCatalogfile.Close()

	b, err := ioutil.ReadAll(dummyCatalogfile)
	Expect(err).ToNot(HaveOccurred())

	nodeCatalog := rackhdapi.NodeCatalog{}

	err = json.Unmarshal(b, &nodeCatalog)
	Expect(err).ToNot(HaveOccurred())
	return nodeCatalog
}

func MakeTryReservationHandlers(requestID string, nodeID string, expectedNodesPath string, expectedNodeCatalogPath string) []http.HandlerFunc {
	expectedNodes := LoadNodes(expectedNodesPath)
	expectedNodesData, err := json.Marshal(expectedNodes)
	Expect(err).ToNot(HaveOccurred())
	var expectedNode rackhdapi.Node
	for n := range expectedNodes {
		if expectedNodes[n].ID == nodeID {
			expectedNode = expectedNodes[n]
		}
	}
	Expect(expectedNode).ToNot(BeNil())
	expectedNodeData, err := json.Marshal(expectedNode)
	Expect(err).ToNot(HaveOccurred())
	expectedNodeCatalog := LoadNodeCatalog(expectedNodeCatalogPath)
	expectedNodeCatalogData, err := json.Marshal(expectedNodeCatalog)
	Expect(err).ToNot(HaveOccurred())

	reservationHandlers := []http.HandlerFunc{
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/common/nodes"),
			ghttp.RespondWith(http.StatusOK, expectedNodesData),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s/catalogs/ohai", nodeID)),
			ghttp.RespondWith(http.StatusOK, expectedNodeCatalogData),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodeID)),
			ghttp.RespondWith(http.StatusOK, nil),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodeID)),
			ghttp.RespondWith(http.StatusOK, expectedNodeData),
		),
	}

	return append(reservationHandlers, MakeWorkflowHandlers("Reserve", requestID, nodeID)...)
}

func MakeWorkflowHandlers(workflow string, requestID string, nodeID string) []http.HandlerFunc {
	taskStubData := []byte(fmt.Sprintf("[{\"injectableName\": \"Task.BOSH.%s.Node.%s\"}]", workflow, requestID))
	workflowStubData := []byte(fmt.Sprintf("[{\"injectableName\": \"Graph.BOSH.%sNode.%s\"}]", workflow, requestID))
	nodeStubData := []byte(`{"obmSettings": [{"service": "fake-obm-service"}]}`)
	completedWorkflowResponse := []byte(fmt.Sprintf("{\"id\": \"%s\", \"_status\": \"succeeded\"}", requestID))

	return []http.HandlerFunc{
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
			ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/nodes/%s", nodeID)),
			ghttp.RespondWith(http.StatusOK, nodeStubData),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("POST", fmt.Sprintf("/api/1.1/nodes/%s/workflows/", nodeID)),
			ghttp.RespondWith(http.StatusCreated, completedWorkflowResponse),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/workflows/%s", requestID)),
			ghttp.RespondWith(http.StatusOK, completedWorkflowResponse),
		),
	}
}
