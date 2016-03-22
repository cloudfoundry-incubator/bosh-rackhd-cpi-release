package rackhdapi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
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

	Describe("GetActiveWorkflows", func() {
		Context("there is a running workflow", func() {
			It("returns a node's active workflow", func() {
				rawWorkflow := helpers.LoadJSON("../spec_assets/dummy_workflow_response.json")
				httpResponse := []byte(fmt.Sprintf("%s", string(rawWorkflow)))
				var expectedResponse rackhdapi.WorkflowResponse
				err := json.Unmarshal(httpResponse, &expectedResponse)
				Expect(err).ToNot(HaveOccurred())

				nodeID := "nodeID"
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodeID)),
						ghttp.RespondWith(http.StatusOK, httpResponse),
					),
				)

				response, err := rackhdapi.GetActiveWorkflows(cpiConfig, nodeID)
				Expect(err).ToNot(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(response).To(Equal(expectedResponse))
			})
		})

		Context("there is no running workflow", func() {
			It("returns nil", func() {
				nodeID := "nodeID"
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("/api/1.1/nodes/%s/workflows/active", nodeID)),
						ghttp.RespondWith(http.StatusNoContent, []byte{}),
					),
				)

				response, err := rackhdapi.GetActiveWorkflows(cpiConfig, nodeID)
				Expect(err).ToNot(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(response).To(Equal(rackhdapi.WorkflowResponse{}))
			})
		})

	})

	Describe("WorkflowFetcher", func() {
		It("returns the workflow with the specified ID", func() {
			httpResponse := helpers.LoadJSON("../spec_assets/dummy_workflow_response.json")
			var expectedResponse rackhdapi.WorkflowResponse
			err := json.Unmarshal(httpResponse, &expectedResponse)
			Expect(err).ToNot(HaveOccurred())

			workflowID := "5665a788fd797bfc044efe6e"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/common/workflows/%s", workflowID)),
					ghttp.RespondWith(http.StatusOK, httpResponse),
				),
			)

			response, err := rackhdapi.WorkflowFetcher(cpiConfig, workflowID)

			Expect(err).ToNot(HaveOccurred())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(response).To(Equal(expectedResponse))
		})
	})

	Describe("PublishWorkflow", func() {
		It("add workflow to library, retrieves updated list of tasks from task library", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			uuidObj, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uuid := uuidObj.String()
			cpiConfig := config.Cpi{ApiServer: apiServer}

			fakeTaskStub := rackhdapi.TaskStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: rackhdapi.DefaultUnusedName,
			}

			fakeTask := struct {
				*rackhdapi.TaskStub
				*rackhdapi.OptionContainer
			}{
				TaskStub:        &fakeTaskStub,
				OptionContainer: &rackhdapi.OptionContainer{},
			}

			fakeTaskBytes, err := json.Marshal(fakeTask)
			Expect(err).ToNot(HaveOccurred())

			err = rackhdapi.PublishTask(cpiConfig, fakeTaskBytes)
			Expect(err).ToNot(HaveOccurred())

			fakeTasks := []rackhdapi.WorkflowTask{
				rackhdapi.WorkflowTask{
					TaskName: "fake-task-name",
					Label:    "fake-label",
					WaitOn: map[string]string{
						"fake-take": "succeeded",
					},
					IgnoreFailure: true,
				},
			}

			fakeWorkflowStub := rackhdapi.WorkflowStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: rackhdapi.DefaultUnusedName,
				Tasks:      fakeTasks,
			}

			fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
			Expect(err).ToNot(HaveOccurred())

			err = rackhdapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
			Expect(err).ToNot(HaveOccurred())

			workflowLibraryBytes, err := rackhdapi.RetrieveWorkflows(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			workflowLibrary := []rackhdapi.WorkflowStub{}
			err = json.Unmarshal(workflowLibraryBytes, &workflowLibrary)
			Expect(err).ToNot(HaveOccurred())

			Expect(workflowLibrary).To(ContainElement(fakeWorkflowStub))
		})
	})

	Describe("RunWorkflow", func() {
		Context("when the workflow completes with valid status", func() {
			It("returns", func() {
				workflowResponseFile, err := os.Open("../spec_assets/dummy_completed_workflow_response.json")
				Expect(err).ToNot(HaveOccurred())
				defer workflowResponseFile.Close()

				workflowResponseBytes, err := ioutil.ReadAll(workflowResponseFile)
				Expect(err).ToNot(HaveOccurred())

				var wr rackhdapi.WorkflowResponse
				err = json.Unmarshal(workflowResponseBytes, &wr)
				Expect(err).ToNot(HaveOccurred())

				fakeWorkflowPoster := func(config.Cpi, string, rackhdapi.RunWorkflowRequestBody) (rackhdapi.WorkflowResponse, error) {
					return wr, nil
				}

				fakeWorkflowFetcher := func(config.Cpi, string) (rackhdapi.WorkflowResponse, error) {
					return wr, nil
				}

				c := config.Cpi{RunWorkflowTimeoutSeconds: 5}
				err = rackhdapi.RunWorkflow(fakeWorkflowPoster, fakeWorkflowFetcher, c, "nodeID", rackhdapi.RunWorkflowRequestBody{})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Describe("INTEGRATION", func() {
			var idleNodes []rackhdapi.Node
			var cpiConfig config.Cpi
			var nodeID string
			var obm string
			var guid string

			type Options struct {
				OBMServiceName *string `json:"obmServiceName"`
			}

			BeforeEach(func() {
				apiServer, err := helpers.GetRackHDHost()
				Expect(err).ToNot(HaveOccurred())
				cpiConfig = config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 2 * 60}

				rejectNodesRunningWorkflows := func(nodes []rackhdapi.Node) []rackhdapi.Node {
					var n []rackhdapi.Node
					for i := range nodes {
						w, err := rackhdapi.GetActiveWorkflows(cpiConfig, nodes[i].ID)
						Expect(err).ToNot(HaveOccurred())
						if w.Name == "" {
							n = append(n, nodes[i])
						}
					}
					return n
				}

				allNodes, err := rackhdapi.GetNodes(cpiConfig)
				Expect(err).ToNot(HaveOccurred())
				idleNodes = rejectNodesRunningWorkflows(allNodes)

				t := time.Now()
				rand.Seed(t.Unix())
				i := rand.Intn(len(idleNodes))
				nodeID = idleNodes[i].ID

				obm, err = rackhdapi.GetOBMServiceName(cpiConfig, nodeID)
				Expect(err).ToNot(HaveOccurred())

				uuidObj, err := uuid.NewV4()
				Expect(err).ToNot(HaveOccurred())
				guid = uuidObj.String()
			})

			Context("when the workflow completes successfully", func() {
				It("returns no error", func() {
					fakeWorkflow := helpers.LoadWorkflow("../spec_assets/dummy_succeeding_workflow.json")
					fakeWorkflow.Name = fmt.Sprintf("Test.Success.CF.Fake.%s", guid)

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflow)
					Expect(err).ToNot(HaveOccurred())

					err = rackhdapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := rackhdapi.RunWorkflowRequestBody{
						Name:    fakeWorkflow.Name,
						Options: map[string]interface{}{"defaults": Options{OBMServiceName: &obm}},
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

					body := rackhdapi.RunWorkflowRequestBody{
						Name:    fakeWorkflow.Name,
						Options: map[string]interface{}{"defaults": Options{OBMServiceName: &obm}},
					}

					err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when the workflow does not complete in the configurable timeout", func() {
				It("returns an error", func() {
					cpiConfig.RunWorkflowTimeoutSeconds = 1

					dummyTask := helpers.LoadTask("../spec_assets/dummy_timeout_task.json")
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

					body := rackhdapi.RunWorkflowRequestBody{
						Name:    fakeWorkflow.Name,
						Options: map[string]interface{}{"defaults": Options{OBMServiceName: &obm}},
					}

					err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).To(HaveOccurred())

					time.Sleep(10 * time.Millisecond)
					Eventually(func() int {
						url := fmt.Sprintf("%s/api/1.1/nodes/%s/workflows/active", cpiConfig.ApiServer, nodeID)
						resp, err := http.Get(url)
						Expect(err).ToNot(HaveOccurred())
						defer resp.Body.Close()

						return resp.StatusCode
					}, 10*time.Second, time.Second).Should(Equal(204))
				})
			})
		})
	})
})
