package rackhdapi_test

import (
  "encoding/json"
  "fmt"
  "net/http"
  "strings"

  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/helpers"
  "github.com/rackhd/rackhd-cpi/models"
  "github.com/rackhd/rackhd-cpi/rackhdapi"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Workflows", func() {
  var server *ghttp.Server
  var jsonReader *strings.Reader
  var cpiConfig config.Cpi

  BeforeEach(func() {
    server, jsonReader, cpiConfig, _ = helpers.SetUp("")
  })

  AfterEach(func() {
    server.Close()
  })

  Describe("GetActiveWorkflow", func() {
    Context("there is a running workflow", func() {
      It("returns a node's active workflow", func() {
        httpResponse := helpers.LoadJSON("../spec_assets/dummy_workflows_response.json")
        var expectedResponse []models.WorkflowResponse
        err := json.Unmarshal(httpResponse, &expectedResponse)
        Expect(err).ToNot(HaveOccurred())

        nodeID := "nodeID"
        server.AppendHandlers(
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodeID), "active=true"),
            ghttp.RespondWith(http.StatusOK, httpResponse),
          ),
        )

        response, err := rackhdapi.GetActiveWorkflow(cpiConfig, nodeID)
        Expect(err).ToNot(HaveOccurred())
        Expect(response).To(Equal(expectedResponse[0]))
      })
    })

    Context("there is no running workflow", func() {
      It("returns nil", func() {
        nodeID := "nodeID"
        server.AppendHandlers(
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s/workflows", nodeID), "active=true"),
            ghttp.RespondWith(http.StatusOK, []byte("[]")),
          ),
        )

        _, err := rackhdapi.GetActiveWorkflow(cpiConfig, nodeID)
        Expect(err).To(HaveOccurred())
      })
    })
  })

  Describe("WorkflowFetcher", func() {
    It("returns the workflow with the specified workflow instance ID", func() {
      httpResponse := helpers.LoadJSON("../spec_assets/dummy_workflow_response.json")
      var expectedResponse models.WorkflowResponse
      err := json.Unmarshal(httpResponse, &expectedResponse)
      Expect(err).ToNot(HaveOccurred())

      workflowID := "5665a788fd797bfc044efe6e"
      url := "/api/2.0/workflows/" + workflowID
      helpers.AddHandler(server, "GET", url, 200, httpResponse)

      response, err := rackhdapi.WorkflowFetcher(cpiConfig, workflowID)
      Expect(err).ToNot(HaveOccurred())
      Expect(response).To(Equal(expectedResponse))
    })
  })

  Describe("PublishWorkflow", func() {
    It("add workflow to library, retrieves updated list of tasks from task library", func() {
      apiServer, err := helpers.GetRackHDHost()
      Expect(err).ToNot(HaveOccurred())
      cpiConfig := config.Cpi{ApiServer: apiServer}

      uuid, err := helpers.GenerateUUID()
      Expect(err).ToNot(HaveOccurred())

      fakeTaskName := "Task.CF.Fake." + uuid
      fakeTask := models.Task{
        Name:           fakeTaskName,
        UnusedName:     models.DefaultUnusedName,
        ImplementsTask: "Task.Base.Linux.Commands",
        Options:        map[string]interface{}{"commands": ""},
      }
      fakeTaskBytes, err := json.Marshal(fakeTask)
      Expect(err).ToNot(HaveOccurred())

      err = rackhdapi.PublishTask(cpiConfig, fakeTaskBytes)
      Expect(err).ToNot(HaveOccurred())

      fakeTasks := []models.WorkflowTask{
        models.WorkflowTask{
          TaskName: fakeTaskName,
          Label:    "fake-label",
        },
      }
      fakeGraphName := "Graph.CF.Fake." + uuid
      fakeGraph := models.Graph{
        Name:       fakeGraphName,
        UnusedName: models.DefaultUnusedName,
        Options:    map[string]interface{}{"commands": "true"},
        Tasks:      fakeTasks,
      }
      fakeGraphBytes, err := json.Marshal(fakeGraph)
      Expect(err).ToNot(HaveOccurred())

      err = rackhdapi.PublishWorkflow(cpiConfig, fakeGraphBytes)
      Expect(err).ToNot(HaveOccurred())

      publishedGraph, err := rackhdapi.RetrieveGraph(cpiConfig, fakeGraphName)
      Expect(err).ToNot(HaveOccurred())
      fakeGraph.Tasks[0].TaskName = "/api/2.0/workflows/tasks/" + fakeTaskName
      Expect(publishedGraph).To(Equal(fakeGraph))
    })
  })

  Describe("INTEGRATION", func() {
    var availableNodes []models.Node
    var cpiConfig config.Cpi
    var nodeID string
    var obm string
    var guid string

    type Options struct {
      NodeID         string  `json:"nodeId"`
      OBMServiceName *string `json:"obmServiceName"`
    }

    BeforeEach(func() {
      apiServer, err := helpers.GetRackHDHost()
      Expect(err).ToNot(HaveOccurred())
      cpiConfig = config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 2 * 60}

      availableNodes, err = rackhdapi.GetAvailableNodes(cpiConfig)
      Expect(err).ToNot(HaveOccurred())
      Expect(len(availableNodes)).To(BeNumerically(">", 0))

      nodeID = availableNodes[0].ID
      obm, err = rackhdapi.GetOBMServiceName(cpiConfig, nodeID)
      Expect(err).ToNot(HaveOccurred())

      guid, err = helpers.GenerateUUID()
      Expect(err).ToNot(HaveOccurred())
    })

    Context("when the workflow completes successfully", func() {
      It("returns no error", func() {
        fakeWorkflow := helpers.LoadWorkflow("../spec_assets/dummy_succeeding_workflow.json")
        fakeWorkflow.Name += guid

        fakeGraphBytes, err := json.Marshal(fakeWorkflow)
        Expect(err).ToNot(HaveOccurred())

        err = rackhdapi.PublishWorkflow(cpiConfig, fakeGraphBytes)
        Expect(err).ToNot(HaveOccurred())

        body := models.RunWorkflowRequestBody{
          Name:    fakeWorkflow.Name,
          Options: map[string]interface{}{"defaults": Options{OBMServiceName: &obm, NodeID: nodeID}},
        }

        err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
        Expect(err).ToNot(HaveOccurred())
      })
    })

    Context("when the workflow completes with failure", func() {
      It("returns an error", func() {
        dummyTask := helpers.LoadTask("../spec_assets/dummy_failing_task.json")
        dummyTask.Name += guid

        dummyTaskBytes, err := json.Marshal(dummyTask)
        Expect(err).ToNot(HaveOccurred())

        err = rackhdapi.PublishTask(cpiConfig, dummyTaskBytes)
        Expect(err).ToNot(HaveOccurred())

        fakeWorkflow := helpers.LoadWorkflow("../spec_assets/dummy_failing_workflow.json")
        fakeWorkflow.Name += guid
        fakeWorkflow.Tasks[3].TaskName += guid

        fakeWorkflowBytes, err := json.Marshal(fakeWorkflow)
        Expect(err).ToNot(HaveOccurred())

        err = rackhdapi.PublishWorkflow(cpiConfig, fakeWorkflowBytes)
        Expect(err).ToNot(HaveOccurred())

        body := models.RunWorkflowRequestBody{
          Name:    fakeWorkflow.Name,
          Options: map[string]interface{}{"defaults": Options{OBMServiceName: &obm}},
        }

        err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
        Expect(err).To(HaveOccurred())
        Expect(err.Error()).To(MatchRegexp(".+failed against node.+"))
      })
    })

    XContext("when the workflow does not complete in the configurable timeout", func() {
      It("returns an error", func() {
        dummyTask := helpers.LoadTask("../spec_assets/dummy_timeout_task.json")
        dummyTask.Name += guid
        dummyTaskBytes, err := json.Marshal(dummyTask)
        Expect(err).ToNot(HaveOccurred())

        err = rackhdapi.PublishTask(cpiConfig, dummyTaskBytes)
        Expect(err).ToNot(HaveOccurred())

        fakeWorkflow := helpers.LoadWorkflow("../spec_assets/dummy_timeout_workflow.json")
        fakeWorkflow.Name += guid
        fakeWorkflow.Tasks[3].TaskName += guid
        fakeWorkflowBytes, err := json.Marshal(fakeWorkflow)
        Expect(err).ToNot(HaveOccurred())

        err = rackhdapi.PublishWorkflow(cpiConfig, fakeWorkflowBytes)
        Expect(err).ToNot(HaveOccurred())

        body := models.RunWorkflowRequestBody{
          Name:    fakeWorkflow.Name,
          Options: map[string]interface{}{"defaults": Options{OBMServiceName: &obm}},
        }
        err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
        Expect(err).To(HaveOccurred())
      })
    })
  })
})
