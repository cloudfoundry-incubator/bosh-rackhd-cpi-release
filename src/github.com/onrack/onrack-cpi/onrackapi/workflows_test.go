package onrackapi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
	"github.com/onrack/onrack-cpi/workflows"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workflows", func() {

	Describe("PublishWorkflow", func() {
		It("add workflow to library, retrieves updated list of tasks from task library", func() {
			apiServer := os.Getenv("ON_RACK_API_URI")
			Expect(apiServer).ToNot(BeEmpty())

			uuidObj, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uuid := uuidObj.String()
			cpiConfig := config.Cpi{ApiServer: apiServer}

			fakeTaskStub := onrackapi.TaskStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackapi.DefaultUnusedName,
			}

			fakeTask := struct {
				*onrackapi.TaskStub
				*onrackapi.OptionContainer
			}{
				TaskStub:        &fakeTaskStub,
				OptionContainer: &onrackapi.OptionContainer{},
			}

			fakeTaskBytes, err := json.Marshal(fakeTask)
			Expect(err).ToNot(HaveOccurred())

			err = onrackapi.PublishTask(cpiConfig, fakeTaskBytes)
			Expect(err).ToNot(HaveOccurred())

			fakeTasks := []onrackapi.WorkflowTask{
				onrackapi.WorkflowTask{
					TaskName: "fake-task-name",
					Label:    "fake-label",
					WaitOn: map[string]string{
						"fake-take": "succeeded",
					},
					IgnoreFailure: true,
				},
			}

			fakeWorkflowStub := onrackapi.WorkflowStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackapi.DefaultUnusedName,
				Tasks:      fakeTasks,
			}

			fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
			Expect(err).ToNot(HaveOccurred())

			err = onrackapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
			Expect(err).ToNot(HaveOccurred())

			workflowLibraryBytes, err := onrackapi.RetrieveWorkflows(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			workflowLibrary := []onrackapi.WorkflowStub{}
			err = json.Unmarshal(workflowLibraryBytes, &workflowLibrary)
			Expect(err).ToNot(HaveOccurred())

			Expect(workflowLibrary).To(ContainElement(fakeWorkflowStub))
		})
	})

	Describe("RunWorkflow", func() {
		Context("when the workflow completes successfully", func() {
			Describe("SLOW_TEST", func() {
				It("returns no error", func() {
					rejectNodesRunningWorkflows := func(nodes []onrackapi.Node) []onrackapi.Node {
						var n []onrackapi.Node
						for i := range nodes {
							if len(nodes[i].Workflows) == 0 {
								n = append(n, nodes[i])
							}
						}
						return n
					}

					apiServer := os.Getenv("ON_RACK_API_URI")
					Expect(apiServer).ToNot(BeEmpty())

					uuidObj, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())
					uuid := uuidObj.String()
					cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 2 * 60}

					allNodes, err := onrackapi.GetNodes(cpiConfig)
					Expect(err).ToNot(HaveOccurred())

					idleNodes := rejectNodesRunningWorkflows(allNodes)
					t := time.Now()
					rand.Seed(t.Unix())

					i := rand.Intn(len(idleNodes))
					nodeID := idleNodes[i].ID

					fakeWorkflowName := fmt.Sprintf("Test.Success.CF.Fake.%s", uuid)
					fakeTasks := []onrackapi.WorkflowTask{
						onrackapi.WorkflowTask{
							TaskName:      workflows.SetPxeBootTaskName,
							Label:         "set-boot-pxe",
							IgnoreFailure: true,
						},
					}

					fakeWorkflowStub := onrackapi.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: onrackapi.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = onrackapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := onrackapi.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = onrackapi.RunWorkflow(onrackapi.WorkflowPoster, onrackapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the workflow completes with failure", func() {
			Describe("SLOW_TEST", func() {
				It("returns an error", func() {
					rejectNodesRunningWorkflows := func(nodes []onrackapi.Node) []onrackapi.Node {
						var n []onrackapi.Node
						for i := range nodes {
							if len(nodes[i].Workflows) == 0 {
								n = append(n, nodes[i])
							}
						}
						return n
					}

					apiServer := os.Getenv("ON_RACK_API_URI")
					Expect(apiServer).ToNot(BeEmpty())

					uuidObj, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())
					uuid := uuidObj.String()
					cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 10 * 60}

					allNodes, err := onrackapi.GetNodes(cpiConfig)
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
						*onrackapi.TaskStub
						*onrackapi.OptionContainer
						*onrackapi.PropertyContainer
					}{}

					err = json.Unmarshal(b, &dummyTask)
					Expect(err).ToNot(HaveOccurred())

					dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Failure.%s", uuid)
					dummyTask.Name = dummyTaskName

					dummyTaskBytes, err := json.Marshal(dummyTask)
					Expect(err).ToNot(HaveOccurred())

					err = onrackapi.PublishTask(cpiConfig, dummyTaskBytes)
					Expect(err).ToNot(HaveOccurred())

					fakeTasks := []onrackapi.WorkflowTask{
						onrackapi.WorkflowTask{
							TaskName: workflows.SetPxeBootTaskName,
							Label:    "set-boot-pxe",
						},
						onrackapi.WorkflowTask{
							TaskName: workflows.RebootNodeTaskName,
							Label:    "reboot",
							WaitOn: map[string]string{
								"set-boot-pxe": "finished",
							},
						},
						onrackapi.WorkflowTask{
							TaskName: workflows.BootstrapUbuntuTaskName,
							Label:    "bootstrap-ubuntu",
							WaitOn: map[string]string{
								"reboot": "succeeded",
							},
						},
						onrackapi.WorkflowTask{
							TaskName: dummyTaskName,
							Label:    "fake-failure-task-label",
							WaitOn: map[string]string{
								"bootstrap-ubuntu": "succeeded",
							},
						},
					}

					fakeWorkflowName := fmt.Sprintf("Test.Failure.CF.Fake.%s", uuid)
					fakeWorkflowStub := onrackapi.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: onrackapi.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = onrackapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := onrackapi.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = onrackapi.RunWorkflow(onrackapi.WorkflowPoster, onrackapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when the workflow does not complete in the configurable timeout", func() {
			Describe("SLOW_TEST", func() {
				It("returns an error", func() {
					rejectNodesRunningWorkflows := func(nodes []onrackapi.Node) []onrackapi.Node {
						var n []onrackapi.Node
						for i := range nodes {
							if len(nodes[i].Workflows) == 0 {
								n = append(n, nodes[i])
							}
						}
						return n
					}

					apiServer := os.Getenv("ON_RACK_API_URI")
					Expect(apiServer).ToNot(BeEmpty())

					uuidObj, err := uuid.NewV4()
					Expect(err).ToNot(HaveOccurred())
					uuid := uuidObj.String()
					cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 1}

					allNodes, err := onrackapi.GetNodes(cpiConfig)
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
						*onrackapi.TaskStub
						*onrackapi.OptionContainer
						*onrackapi.PropertyContainer
					}{}

					err = json.Unmarshal(b, &dummyTask)
					Expect(err).ToNot(HaveOccurred())

					dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Timeout.%s", uuid)
					dummyTask.Name = dummyTaskName

					dummyTaskBytes, err := json.Marshal(dummyTask)
					Expect(err).ToNot(HaveOccurred())

					err = onrackapi.PublishTask(cpiConfig, dummyTaskBytes)
					Expect(err).ToNot(HaveOccurred())

					fakeTasks := []onrackapi.WorkflowTask{
						onrackapi.WorkflowTask{
							TaskName: workflows.SetPxeBootTaskName,
							Label:    "set-boot-pxe",
						},
						onrackapi.WorkflowTask{
							TaskName: workflows.RebootNodeTaskName,
							Label:    "reboot",
							WaitOn: map[string]string{
								"set-boot-pxe": "finished",
							},
						},
						onrackapi.WorkflowTask{
							TaskName: workflows.BootstrapUbuntuTaskName,
							Label:    "bootstrap-ubuntu",
							WaitOn: map[string]string{
								"reboot": "succeeded",
							},
						},
						onrackapi.WorkflowTask{
							TaskName: dummyTaskName,
							Label:    "fake-timeout-task-label",
							WaitOn: map[string]string{
								"bootstrap-ubuntu": "succeeded",
							},
						},
					}

					fakeWorkflowName := fmt.Sprintf("Test.Timeout.CF.Fake.%s", uuid)
					fakeWorkflowStub := onrackapi.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: onrackapi.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = onrackapi.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := onrackapi.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = onrackapi.RunWorkflow(onrackapi.WorkflowPoster, onrackapi.WorkflowFetcher, cpiConfig, nodeID, body)
					Expect(err).To(HaveOccurred())

					Eventually(func() int {
						url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/active", cpiConfig.ApiServer, nodeID)
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

					var wr onrackapi.WorkflowResponse
					err = json.Unmarshal(workflowResponseBytes, &wr)
					Expect(err).ToNot(HaveOccurred())

					fakeWorkflowPoster := func(config.Cpi, string, onrackapi.RunWorkflowRequestBody) (onrackapi.WorkflowResponse, error) {
						return wr, nil
					}

					fakeWorkflowFetcher := func(config.Cpi, string, string) (onrackapi.WorkflowResponse, error) {
						return wr, nil
					}

					c := config.Cpi{RunWorkflowTimeoutSeconds: 5}
					err = onrackapi.RunWorkflow(fakeWorkflowPoster, fakeWorkflowFetcher, c, "nodeID", onrackapi.RunWorkflowRequestBody{})
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
