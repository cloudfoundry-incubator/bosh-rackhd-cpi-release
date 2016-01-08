package rackhdapi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
	"github.com/rackhd/rackhd-cpi/workflows"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func loadWorkflowsResponse(assetPath string) []byte {
	dummyResponseFile, err := os.Open(assetPath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyResponseFile.Close()

	workflowsResponse, err := ioutil.ReadAll(dummyResponseFile)
	Expect(err).ToNot(HaveOccurred())

	return workflowsResponse
}

var _ = Describe("Workflows", func() {
	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost", "disks":{"system":"/dev/sda"}}, "max_create_vm_attempts":1}`, serverURL.Host))
		cpiConfig, err = config.New(jsonReader)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GetActiveWorkflows", func() {
		Context("there is a running workflow", func() {
			It("returns a node's active workflow", func() {
				rawWorkflow := loadWorkflowsResponse("../spec_assets/dummy_workflow_response.json")
				httpResponse := []byte(fmt.Sprintf("[%s]", string(rawWorkflow)))
				var expectedResponse []rackhdapi.WorkflowResponse
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
				Expect(response).To(BeNil())
			})
		})

	})

	Describe("WorkflowFetcher", func() {

		It("returns the workflow with the specified ID", func() {
			httpResponse := loadWorkflowsResponse("../spec_assets/dummy_workflow_response.json")
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
			apiServer := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
			Expect(apiServer).ToNot(BeEmpty())

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
		Context("when the workflow completes successfully", func() {
			Describe("SLOW_TEST", func() {
				It("returns no error", func() {
					rejectNodesRunningWorkflows := func(nodes []rackhdapi.Node) []rackhdapi.Node {
						var n []rackhdapi.Node
						for i := range nodes {
							if len(nodes[i].Workflows) == 0 {
								n = append(n, nodes[i])
							}
						}
						return n
					}

					apiServer := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
					Expect(apiServer).ToNot(BeEmpty())

					uuidObj, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())
					uuid := uuidObj.String()
					cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 2 * 60}

					allNodes, err := rackhdapi.GetNodes(cpiConfig)
					Expect(err).ToNot(HaveOccurred())

					idleNodes := rejectNodesRunningWorkflows(allNodes)
					t := time.Now()
					rand.Seed(t.Unix())

					i := rand.Intn(len(idleNodes))
					nodeID := idleNodes[i].ID

					fakeWorkflowName := fmt.Sprintf("Test.Success.CF.Fake.%s", uuid)
					fakeTasks := []rackhdapi.WorkflowTask{
						rackhdapi.WorkflowTask{
							TaskName:      workflows.SetPxeBootTaskName,
							Label:         "set-boot-pxe",
							IgnoreFailure: true,
						},
					}

					fakeWorkflowStub := rackhdapi.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: rackhdapi.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = rackhdapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := rackhdapi.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the workflow completes with failure", func() {
			Describe("SLOW_TEST", func() {
				It("returns an error", func() {
					rejectNodesRunningWorkflows := func(nodes []rackhdapi.Node) []rackhdapi.Node {
						var n []rackhdapi.Node
						for i := range nodes {
							if len(nodes[i].Workflows) == 0 {
								n = append(n, nodes[i])
							}
						}
						return n
					}

					apiServer := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
					Expect(apiServer).ToNot(BeEmpty())

					uuidObj, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())
					uuid := uuidObj.String()
					cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 10 * 60}

					allNodes, err := rackhdapi.GetNodes(cpiConfig)
					Expect(err).ToNot(HaveOccurred())

					idleNodes := rejectNodesRunningWorkflows(allNodes)
					t := time.Now()
					rand.Seed(t.Unix())

					i := rand.Intn(len(idleNodes))
					nodeID := idleNodes[i].ID

					dummyTaskFile, err := os.Open("../spec_assets/dummy_failing_task.json")
					Expect(err).ToNot(HaveOccurred())
					defer dummyTaskFile.Close()

					b, err := ioutil.ReadAll(dummyTaskFile)
					Expect(err).ToNot(HaveOccurred())

					dummyTask := struct {
						*rackhdapi.TaskStub
						*rackhdapi.OptionContainer
						*rackhdapi.PropertyContainer
					}{}

					err = json.Unmarshal(b, &dummyTask)
					Expect(err).ToNot(HaveOccurred())

					dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Failure.%s", uuid)
					dummyTask.Name = dummyTaskName

					dummyTaskBytes, err := json.Marshal(dummyTask)
					Expect(err).ToNot(HaveOccurred())

					err = rackhdapi.PublishTask(cpiConfig, dummyTaskBytes)
					Expect(err).ToNot(HaveOccurred())

					fakeTasks := []rackhdapi.WorkflowTask{
						rackhdapi.WorkflowTask{
							TaskName: workflows.SetPxeBootTaskName,
							Label:    "set-boot-pxe",
						},
						rackhdapi.WorkflowTask{
							TaskName: workflows.RebootNodeTaskName,
							Label:    "reboot",
							WaitOn: map[string]string{
								"set-boot-pxe": "finished",
							},
						},
						rackhdapi.WorkflowTask{
							TaskName: workflows.BootstrapUbuntuTaskName,
							Label:    "bootstrap-ubuntu",
							WaitOn: map[string]string{
								"reboot": "succeeded",
							},
						},
						rackhdapi.WorkflowTask{
							TaskName: dummyTaskName,
							Label:    "fake-failure-task-label",
							WaitOn: map[string]string{
								"bootstrap-ubuntu": "succeeded",
							},
						},
					}

					fakeWorkflowName := fmt.Sprintf("Test.Failure.CF.Fake.%s", uuid)
					fakeWorkflowStub := rackhdapi.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: rackhdapi.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = rackhdapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := rackhdapi.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when the workflow does not complete in the configurable timeout", func() {
			Describe("SLOW_TEST", func() {
				It("returns an error", func() {
					rejectNodesRunningWorkflows := func(nodes []rackhdapi.Node) []rackhdapi.Node {
						var n []rackhdapi.Node
						for i := range nodes {
							if len(nodes[i].Workflows) == 0 {
								n = append(n, nodes[i])
							}
						}
						return n
					}

					apiServer := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
					Expect(apiServer).ToNot(BeEmpty())

					uuidObj, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())
					uuid := uuidObj.String()
					cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 1}

					allNodes, err := rackhdapi.GetNodes(cpiConfig)
					Expect(err).ToNot(HaveOccurred())

					idleNodes := rejectNodesRunningWorkflows(allNodes)
					t := time.Now()
					rand.Seed(t.Unix())

					i := rand.Intn(len(idleNodes))
					nodeID := idleNodes[i].ID

					dummyTaskFile, err := os.Open("../spec_assets/dummy_timeout_task.json")
					Expect(err).ToNot(HaveOccurred())
					defer dummyTaskFile.Close()

					b, err := ioutil.ReadAll(dummyTaskFile)
					Expect(err).ToNot(HaveOccurred())

					dummyTask := struct {
						*rackhdapi.TaskStub
						*rackhdapi.OptionContainer
						*rackhdapi.PropertyContainer
					}{}

					err = json.Unmarshal(b, &dummyTask)
					Expect(err).ToNot(HaveOccurred())

					dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Timeout.%s", uuid)
					dummyTask.Name = dummyTaskName

					dummyTaskBytes, err := json.Marshal(dummyTask)
					Expect(err).ToNot(HaveOccurred())

					err = rackhdapi.PublishTask(cpiConfig, dummyTaskBytes)
					Expect(err).ToNot(HaveOccurred())

					fakeTasks := []rackhdapi.WorkflowTask{
						rackhdapi.WorkflowTask{
							TaskName: workflows.SetPxeBootTaskName,
							Label:    "set-boot-pxe",
						},
						rackhdapi.WorkflowTask{
							TaskName: workflows.RebootNodeTaskName,
							Label:    "reboot",
							WaitOn: map[string]string{
								"set-boot-pxe": "finished",
							},
						},
						rackhdapi.WorkflowTask{
							TaskName: workflows.BootstrapUbuntuTaskName,
							Label:    "bootstrap-ubuntu",
							WaitOn: map[string]string{
								"reboot": "succeeded",
							},
						},
						rackhdapi.WorkflowTask{
							TaskName: dummyTaskName,
							Label:    "fake-timeout-task-label",
							WaitOn: map[string]string{
								"bootstrap-ubuntu": "succeeded",
							},
						},
					}

					fakeWorkflowName := fmt.Sprintf("Test.Timeout.CF.Fake.%s", uuid)
					fakeWorkflowStub := rackhdapi.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: rackhdapi.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = rackhdapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := rackhdapi.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).To(HaveOccurred())

					Eventually(func() int {
						url := fmt.Sprintf("http://%s/api/1.1/nodes/%s/workflows/active", cpiConfig.ApiServer, nodeID)
						resp, err := http.Get(url)
						Expect(err).ToNot(HaveOccurred())
						defer resp.Body.Close()

						return resp.StatusCode
					}, 10*time.Second, time.Second).Should(Equal(204))
				})
			})
		})

		Context("when the workflow completes with valid status", func() {
			Describe("SLOW_TEST", func() {
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
		})
	})
})
