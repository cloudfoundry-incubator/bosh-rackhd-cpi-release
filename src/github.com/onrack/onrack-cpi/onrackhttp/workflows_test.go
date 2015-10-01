package onrackhttp_test

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
	"github.com/onrack/onrack-cpi/onrackhttp"
	"github.com/onrack/onrack-cpi/workflows"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workflows", func() {

	Describe("Publishing workflow", func() {
		It("add workflow to library, retrieves updated list of tasks from task library", func() {
			apiServer := os.Getenv("ON_RACK_API_URI")
			Expect(apiServer).ToNot(BeEmpty())

			uuidObj, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uuid := uuidObj.String()
			cpiConfig := config.Cpi{ApiServer: apiServer}

			fakeTaskStub := onrackhttp.TaskStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackhttp.DefaultUnusedName,
			}

			fakeTask := struct {
				*onrackhttp.TaskStub
				*onrackhttp.OptionContainer
			}{
				TaskStub:        &fakeTaskStub,
				OptionContainer: &onrackhttp.OptionContainer{},
			}

			fakeTaskBytes, err := json.Marshal(fakeTask)
			Expect(err).ToNot(HaveOccurred())

			err = onrackhttp.PublishTask(cpiConfig, fakeTaskBytes)
			Expect(err).ToNot(HaveOccurred())

			fakeTasks := []onrackhttp.WorkflowTask{
				onrackhttp.WorkflowTask{
					TaskName: "fake-task-name",
					Label:    "fake-label",
					WaitOn: map[string]string{
						"fake-take": "succeeded",
					},
					IgnoreFailure: true,
				},
			}

			fakeWorkflowStub := onrackhttp.WorkflowStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackhttp.DefaultUnusedName,
				Tasks:      fakeTasks,
			}

			fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
			Expect(err).ToNot(HaveOccurred())

			err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
			Expect(err).ToNot(HaveOccurred())

			workflowLibraryBytes, err := onrackhttp.RetrieveWorkflows(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			workflowLibrary := []onrackhttp.WorkflowStub{}
			err = json.Unmarshal(workflowLibraryBytes, &workflowLibrary)
			Expect(err).ToNot(HaveOccurred())

			Expect(workflowLibrary).To(ContainElement(fakeWorkflowStub))
		})
	})

	Describe("RunWorkflow", func() {
		Context("when the workflow completes successfully", func() {
			Describe("SLOW_TEST", func() {
				It("returns no error", func() {
					rejectNodesRunningWorkflows := func(nodes []onrackhttp.Node) []onrackhttp.Node {
						var n []onrackhttp.Node
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

					allNodes, err := onrackhttp.GetNodes(cpiConfig)
					Expect(err).ToNot(HaveOccurred())

					idleNodes := rejectNodesRunningWorkflows(allNodes)
					t := time.Now()
					rand.Seed(t.Unix())

					i := rand.Intn(len(idleNodes))
					nodeID := idleNodes[i].ID

					fakeWorkflowName := fmt.Sprintf("Test.Success.CF.Fake.%s", uuid)
					fakeTasks := []onrackhttp.WorkflowTask{
						onrackhttp.WorkflowTask{
							TaskName:      workflows.SetPxeBootTaskName,
							Label:         "set-boot-pxe",
							IgnoreFailure: true,
						},
					}

					fakeWorkflowStub := onrackhttp.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: onrackhttp.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := onrackhttp.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = onrackhttp.RunWorkflow(cpiConfig, nodeID, body)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the workflow completes with failure", func() {
			Describe("SLOW_TEST", func() {
				It("returns an error", func() {
					rejectNodesRunningWorkflows := func(nodes []onrackhttp.Node) []onrackhttp.Node {
						var n []onrackhttp.Node
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

					allNodes, err := onrackhttp.GetNodes(cpiConfig)
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
						*onrackhttp.TaskStub
						*onrackhttp.OptionContainer
						*onrackhttp.PropertyContainer
					}{}

					err = json.Unmarshal(b, &dummyTask)
					Expect(err).ToNot(HaveOccurred())

					dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Failure.%s", uuid)
					dummyTask.Name = dummyTaskName

					dummyTaskBytes, err := json.Marshal(dummyTask)
					Expect(err).ToNot(HaveOccurred())

					err = onrackhttp.PublishTask(cpiConfig, dummyTaskBytes)
					Expect(err).ToNot(HaveOccurred())

					fakeTasks := []onrackhttp.WorkflowTask{
						onrackhttp.WorkflowTask{
							TaskName: workflows.SetPxeBootTaskName,
							Label:    "set-boot-pxe",
						},
						onrackhttp.WorkflowTask{
							TaskName: workflows.RebootNodeTaskName,
							Label:    "reboot",
							WaitOn: map[string]string{
								"set-boot-pxe": "finished",
							},
						},
						onrackhttp.WorkflowTask{
							TaskName: workflows.BootstrapUbuntuTaskName,
							Label:    "bootstrap-ubuntu",
							WaitOn: map[string]string{
								"reboot": "succeeded",
							},
						},
						onrackhttp.WorkflowTask{
							TaskName: dummyTaskName,
							Label:    "fake-failure-task-label",
							WaitOn: map[string]string{
								"bootstrap-ubuntu": "succeeded",
							},
						},
					}

					fakeWorkflowName := fmt.Sprintf("Test.Failure.CF.Fake.%s", uuid)
					fakeWorkflowStub := onrackhttp.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: onrackhttp.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := onrackhttp.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = onrackhttp.RunWorkflow(cpiConfig, nodeID, body)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when the workflow does not complete in the configurable timeout", func() {
			Describe("SLOW_TEST", func() {
				It("returns an error", func() {
					rejectNodesRunningWorkflows := func(nodes []onrackhttp.Node) []onrackhttp.Node {
						var n []onrackhttp.Node
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

					allNodes, err := onrackhttp.GetNodes(cpiConfig)
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
						*onrackhttp.TaskStub
						*onrackhttp.OptionContainer
						*onrackhttp.PropertyContainer
					}{}

					err = json.Unmarshal(b, &dummyTask)
					Expect(err).ToNot(HaveOccurred())

					dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Timeout.%s", uuid)
					dummyTask.Name = dummyTaskName

					dummyTaskBytes, err := json.Marshal(dummyTask)
					Expect(err).ToNot(HaveOccurred())

					err = onrackhttp.PublishTask(cpiConfig, dummyTaskBytes)
					Expect(err).ToNot(HaveOccurred())

					fakeTasks := []onrackhttp.WorkflowTask{
						onrackhttp.WorkflowTask{
							TaskName: workflows.SetPxeBootTaskName,
							Label:    "set-boot-pxe",
						},
						onrackhttp.WorkflowTask{
							TaskName: workflows.RebootNodeTaskName,
							Label:    "reboot",
							WaitOn: map[string]string{
								"set-boot-pxe": "finished",
							},
						},
						onrackhttp.WorkflowTask{
							TaskName: workflows.BootstrapUbuntuTaskName,
							Label:    "bootstrap-ubuntu",
							WaitOn: map[string]string{
								"reboot": "succeeded",
							},
						},
						onrackhttp.WorkflowTask{
							TaskName: dummyTaskName,
							Label:    "fake-timeout-task-label",
							WaitOn: map[string]string{
								"bootstrap-ubuntu": "succeeded",
							},
						},
					}

					fakeWorkflowName := fmt.Sprintf("Test.Timeout.CF.Fake.%s", uuid)
					fakeWorkflowStub := onrackhttp.WorkflowStub{
						Name:       fakeWorkflowName,
						UnusedName: onrackhttp.DefaultUnusedName,
						Tasks:      fakeTasks,
					}

					fakeWorkflowStubBytes, err := json.Marshal(fakeWorkflowStub)
					Expect(err).ToNot(HaveOccurred())

					err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflowStubBytes)
					Expect(err).ToNot(HaveOccurred())

					body := onrackhttp.RunWorkflowRequestBody{
						Name: fakeWorkflowName,
					}

					err = onrackhttp.RunWorkflow(cpiConfig, nodeID, body)
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
	})
})
