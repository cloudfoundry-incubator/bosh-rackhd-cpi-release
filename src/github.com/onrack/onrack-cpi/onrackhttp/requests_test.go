package onrackhttp_test

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
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
	"github.com/onrack/onrack-cpi/workflows"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Requests", func() {
	Describe("uploading to then deleting from the OnRack API", func() {
		It("allows files to be uploaded and deleted", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}
			dummyStr := "Some ice cold file"
			dummyFile := strings.NewReader(dummyStr)

			uuid, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())

			url := fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", c.ApiServer, uuid.String())
			resp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))

			onrackUUID, err := onrackhttp.UploadFile(c, uuid.String(), dummyFile, int64(len(dummyStr)))
			Expect(err).ToNot(HaveOccurred())
			Expect(onrackUUID).ToNot(BeEmpty())

			url = fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", c.ApiServer, uuid.String())
			getResp, err := http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			respBytes, err := ioutil.ReadAll(getResp.Body)
			Expect(err).ToNot(HaveOccurred())

			defer getResp.Body.Close()
			Expect(getResp.StatusCode).To(Equal(200))

			fileMetadataResp := onrackhttp.FileMetadataResponse{}
			err = json.Unmarshal(respBytes, &fileMetadataResp)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileMetadataResp).To(HaveLen(1))

			fileMetadata := fileMetadataResp[0]
			Expect(fileMetadata.Basename).To(Equal(uuid.String()))

			err = onrackhttp.DeleteFile(c, onrackUUID)
			Expect(err).ToNot(HaveOccurred())

			url = fmt.Sprintf("http://%s:8080/api/common/files/metadata/%s", c.ApiServer, uuid.String())
			resp, err = http.Get(url)
			Expect(err).ToNot(HaveOccurred())

			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})
	})

	Describe("Getting nodes", func() {
		It("return expected nodes fields", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}

			nodes, err := onrackhttp.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())

			resp, err := http.Get(fmt.Sprintf("http://%s:8080/api/common/nodes", c.ApiServer))
			Expect(err).ToNot(HaveOccurred())

			b, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var rawNodes []onrackhttp.Node
			err = json.Unmarshal(b, &rawNodes)
			Expect(err).ToNot(HaveOccurred())

			Expect(nodes).To(Equal(rawNodes))
		})
	})

	Describe("Getting catalog", func() {
		It("return ", func() {
			apiServerIP := os.Getenv("ON_RACK_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}

			nodes, err := onrackhttp.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())

			Expect(nodes).ToNot(BeEmpty())
			testNode := nodes[0]

			catalog, err := onrackhttp.GetNodeCatalog(c, testNode.ID)
			Expect(err).ToNot(HaveOccurred())

			resp, err := http.Get(fmt.Sprintf("http://%s:8080/api/common/nodes/%s/catalogs/ohai", c.ApiServer, testNode.ID))
			Expect(err).ToNot(HaveOccurred())

			b, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var rawCatalog onrackhttp.NodeCatalog
			err = json.Unmarshal(b, &rawCatalog)
			Expect(err).ToNot(HaveOccurred())

			Expect(catalog).To(Equal(rawCatalog))
		})
	})

	Describe("Publishing tasks", func() {
		It("adds task to library, retrieves updated list of tasks from task library", func() {
			apiServer := os.Getenv("ON_RACK_API_URI")
			Expect(apiServer).ToNot(BeEmpty())

			uuidObj, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uuid := uuidObj.String()
			cpiConfig := config.Cpi{ApiServer: apiServer}

			fakeTask := onrackhttp.Task{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackhttp.DefaultUnusedName,
				Options: map[string]interface{}{
					"option_1": "foo",
				},
			}

			err = onrackhttp.PublishTask(cpiConfig, fakeTask)
			Expect(err).ToNot(HaveOccurred())

			taskLibrary, err := onrackhttp.RetrieveTasks(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			Expect(taskLibrary).To(ContainElement(fakeTask))
		})
	})

	Describe("Publishing workflow", func() {
		It("add workflow to library, retrieves updated list of tasks from task library", func() {
			apiServer := os.Getenv("ON_RACK_API_URI")
			Expect(apiServer).ToNot(BeEmpty())

			uuidObj, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uuid := uuidObj.String()
			cpiConfig := config.Cpi{ApiServer: apiServer}

			fakeTask := onrackhttp.Task{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackhttp.DefaultUnusedName,
				Options: map[string]interface{}{
					"option_1": "foo",
				},
			}

			err = onrackhttp.PublishTask(cpiConfig, fakeTask)
			Expect(err).ToNot(HaveOccurred())

			fakeWorkflow := onrackhttp.Workflow{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackhttp.DefaultUnusedName,
				Tasks: []onrackhttp.WorkflowTask{
					onrackhttp.WorkflowTask{
						TaskName: "fake-task-name",
						Label:    "fake-label",
						WaitOn: map[string]string{
							"fake-take": "succeeded",
						},
						IgnoreFailure: true,
					},
				},
			}

			err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflow)
			Expect(err).ToNot(HaveOccurred())

			workflowLibrary, err := onrackhttp.RetrieveWorkflows(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			for i := range workflowLibrary {
				delete(workflowLibrary[i].Options, "defaults")
				delete(workflowLibrary[i].Options, "bootstrap-ubuntu")
			}
			Expect(workflowLibrary).To(ContainElement(fakeWorkflow))
		})
	})

	Describe("RunWorkflow", func() {
		Context("when the workflow completes successfully", func() {
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
				cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 10 * 60}

				allNodes, err := onrackhttp.GetNodes(cpiConfig)
				Expect(err).ToNot(HaveOccurred())

				idleNodes := rejectNodesRunningWorkflows(allNodes)
				t := time.Now()
				rand.Seed(t.Unix())

				i := rand.Intn(len(idleNodes))
				nodeID := idleNodes[i].ID

				dummyTaskFile, err := os.Open("../spec_assets/dummy_succeeding_task.json")
				Expect(err).ToNot(HaveOccurred())
				defer dummyTaskFile.Close()

				b, err := ioutil.ReadAll(dummyTaskFile)
				Expect(err).ToNot(HaveOccurred())

				dummyTask := onrackhttp.Task{}

				err = json.Unmarshal(b, &dummyTask)
				Expect(err).ToNot(HaveOccurred())

				dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Success.%s", uuid)
				dummyTask.Name = dummyTaskName

				err = onrackhttp.PublishTask(cpiConfig, dummyTask)
				Expect(err).ToNot(HaveOccurred())

				fakeWorkflowName := fmt.Sprintf("Test.Success.CF.Fake.%s", uuid)
				fakeWorkflow := onrackhttp.Workflow{
					Name:       fakeWorkflowName,
					UnusedName: onrackhttp.DefaultUnusedName,
					Tasks: []onrackhttp.WorkflowTask{
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
							Label:    "fake-success-task-label",
							WaitOn: map[string]string{
								"bootstrap-ubuntu": "succeeded",
							},
						},
					},
				}
				err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflow)
				Expect(err).ToNot(HaveOccurred())

				body := onrackhttp.RunWorkflowRequestBody{
					Name: fakeWorkflowName,
				}

				err = onrackhttp.RunWorkflow(cpiConfig, nodeID, body)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the workflow completes with failure", func() {
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

				dummyTask := onrackhttp.Task{}

				err = json.Unmarshal(b, &dummyTask)
				Expect(err).ToNot(HaveOccurred())

				dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Failure.%s", uuid)
				dummyTask.Name = dummyTaskName

				err = onrackhttp.PublishTask(cpiConfig, dummyTask)
				Expect(err).ToNot(HaveOccurred())

				fakeWorkflowName := fmt.Sprintf("Test.Failure.CF.Fake.%s", uuid)
				fakeWorkflow := onrackhttp.Workflow{
					Name:       fakeWorkflowName,
					UnusedName: onrackhttp.DefaultUnusedName,
					Tasks: []onrackhttp.WorkflowTask{
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
					},
				}
				err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflow)
				Expect(err).ToNot(HaveOccurred())

				body := onrackhttp.RunWorkflowRequestBody{
					Name: fakeWorkflowName,
				}

				err = onrackhttp.RunWorkflow(cpiConfig, nodeID, body)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the workflow does not complete in the configurable timeout", func() {
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
				cpiConfig := config.Cpi{ApiServer: apiServer, RunWorkflowTimeoutSeconds: 60}

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

				dummyTask := onrackhttp.Task{}

				err = json.Unmarshal(b, &dummyTask)
				Expect(err).ToNot(HaveOccurred())

				dummyTaskName := fmt.Sprintf("Requests.Test.Dummy.Timeout.%s", uuid)
				dummyTask.Name = dummyTaskName

				err = onrackhttp.PublishTask(cpiConfig, dummyTask)
				Expect(err).ToNot(HaveOccurred())

				fakeWorkflowName := fmt.Sprintf("Test.Timeout.CF.Fake.%s", uuid)
				fakeWorkflow := onrackhttp.Workflow{
					Name:       fakeWorkflowName,
					UnusedName: onrackhttp.DefaultUnusedName,
					Tasks: []onrackhttp.WorkflowTask{
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
					},
				}
				err = onrackhttp.PublishWorkflow(cpiConfig, fakeWorkflow)
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
