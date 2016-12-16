package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/models"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

const (
	env_rackhd_api_url = "RACKHD_API_URL"
)

func AddHandler(server *ghttp.Server, method, url string, statusCode int, respBody []byte) {
	server.AppendHandlers(
		ghttp.CombineHandlers(
			ghttp.VerifyRequest(method, url),
			ghttp.RespondWith(statusCode, respBody),
		),
	)
}

func AddHandlerWithParam(server *ghttp.Server, method, url, param string, statusCode int, respBody []byte) {
	server.AppendHandlers(
		ghttp.CombineHandlers(
			ghttp.VerifyRequest(method, url, param),
			ghttp.RespondWith(statusCode, respBody),
		),
	)
}

func SetUp(cpiRequestType string) (*ghttp.Server, *strings.Reader, config.Cpi, bosh.CpiRequest) {
	var err error
	server := ghttp.NewServer()
	jsonReader := strings.NewReader(fmt.Sprintf(`{"api_url":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_reserve_node_attempts":1}`, server.URL()))
	request := bosh.CpiRequest{Method: cpiRequestType}
	cpiConfig, err := config.New(jsonReader, request)
	Expect(err).ToNot(HaveOccurred())

	return server, jsonReader, cpiConfig, request
}

func GetRackHDHost() (string, error) {
	raw_url := os.Getenv(env_rackhd_api_url)
	if raw_url == "" {
		return "", fmt.Errorf("Environment variable %s not set", env_rackhd_api_url)
	}
	return raw_url, nil
}

func LoadJSON(nodePath string) []byte {
	dummyResponseFile, err := os.Open(nodePath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyResponseFile.Close()

	dummyResponseBytes, err := ioutil.ReadAll(dummyResponseFile)
	Expect(err).ToNot(HaveOccurred())

	return dummyResponseBytes
}

func LoadStruct(filePath string, o interface{}) interface{} {
	dummyResponseBytes := LoadJSON(filePath)

	err := json.Unmarshal(dummyResponseBytes, o)
	Expect(err).ToNot(HaveOccurred())

	return o
}

func LoadWorkflow(workflowPath string) models.Workflow {
	workflow := models.Workflow{}
	return *LoadStruct(workflowPath, &workflow).(*models.Workflow)
}

func LoadTask(taskPath string) models.Task {
	task := models.Task{}
	return *LoadStruct(taskPath, &task).(*models.Task)
}

func LoadNodes(nodePath string) []models.Node {
	dummyResponseBytes := LoadJSON(nodePath)

	nodes := []models.Node{}
	err := json.Unmarshal(dummyResponseBytes, &nodes)
	Expect(err).ToNot(HaveOccurred())

	return nodes
}

func LoadNode(nodePath string) models.Node {
	node := models.Node{}
	return *LoadStruct(nodePath, &node).(*models.Node)
}

func LoadTagNodes(nodePath string) []models.TagNode {
	dummyResponseBytes := LoadJSON(nodePath)

	nodes := []models.TagNode{}
	err := json.Unmarshal(dummyResponseBytes, &nodes)
	Expect(err).ToNot(HaveOccurred())

	return nodes
}

func LoadTagNode(nodePath string) models.TagNode {
	node := models.TagNode{}
	return *LoadStruct(nodePath, &node).(*models.TagNode)
}

func LoadNodeCatalog(nodeCatalogPath string) models.NodeCatalog {
	nodeCatalog := models.NodeCatalog{}
	return *LoadStruct(nodeCatalogPath, &nodeCatalog).(*models.NodeCatalog)
}

func MakeTryReservationHandlers(requestID string, nodeID string) []http.HandlerFunc {
	allNodesBytes := LoadJSON("../spec_assets/nodes_all.json")
	unavailableNodesBytes := LoadJSON("../spec_assets/tag_nodes_reserved.json")
	blockedNodesBytes := LoadJSON("../spec_assets/tag_nodes_blocked.json")

	reservationHandlers := []http.HandlerFunc{
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/2.0/tags/unavailable/nodes"),
			ghttp.RespondWith(http.StatusOK, unavailableNodesBytes),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/2.0/tags/blocked/nodes"),
			ghttp.RespondWith(http.StatusOK, blockedNodesBytes),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/2.0/nodes", "type=compute"),
			ghttp.RespondWith(http.StatusOK, allNodesBytes),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/2.0/nodes/"+nodeID+"/workflows", "active=true"),
			ghttp.RespondWith(http.StatusOK, []byte("[]")),
		),
	}

	return append(reservationHandlers, MakeWorkflowHandlers("Reserve", requestID, nodeID)...)
}

func MakeWorkflowHandlers(workflow string, requestID string, nodeID string) []http.HandlerFunc {
	taskName := fmt.Sprintf("Task.BOSH.%s.Node.%s", workflow, requestID)
	taskBytes := []byte(fmt.Sprintf("{\"injectableName\": \"%s\"}", taskName))
	taskArrayBytes := []byte("[" + string(taskBytes) + "]")

	graphName := fmt.Sprintf("Graph.BOSH.Node.%s.%s", workflow, requestID)
	graphBytes := []byte(fmt.Sprintf("{\"injectableName\": \"%s\"}", graphName))
	graphArrayBytes := []byte("[" + string(graphBytes) + "]")

	nodeOBMBytes := []byte(`{"obms": [{"service": "fake-obm-service"}]}`)
	completedWorkflowResponse := []byte(fmt.Sprintf("{\"instanceId\": \"%s\", \"status\": \"succeeded\"}", requestID))

	return []http.HandlerFunc{
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("PUT", "/api/2.0/workflows/tasks"),
			ghttp.RespondWith(http.StatusCreated, taskBytes),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/2.0/workflows/tasks/"+taskName),
			ghttp.RespondWith(http.StatusOK, taskArrayBytes),
		),

		ghttp.CombineHandlers(
			ghttp.VerifyRequest("PUT", "/api/2.0/workflows/graphs"),
			ghttp.RespondWith(http.StatusCreated, graphBytes),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/api/2.0/workflows/graphs/"+graphName),
			ghttp.RespondWith(http.StatusOK, graphArrayBytes),
		),

		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
			ghttp.RespondWith(http.StatusOK, nodeOBMBytes),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("POST", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodeID)),
			ghttp.RespondWith(http.StatusCreated, completedWorkflowResponse),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/workflows/%s", requestID)),
			ghttp.RespondWith(http.StatusOK, completedWorkflowResponse),
		),
	}
}
