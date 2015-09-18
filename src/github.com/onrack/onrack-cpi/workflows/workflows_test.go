package workflows_test

import (
	"encoding/json"
	"fmt"
	"github.com/onrack/onrack-cpi/workflows"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"

	"io/ioutil"
	"net/http"
	"github.com/nu7hatch/gouuid"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workflows", func() {
	Describe("PublishCreateVMWorkflow", func (){
		It("returns", func(){
			apiServer := os.Getenv("ON_RACK_API_URI")
			cpiConfig := config.Cpi{ApiServer: apiServer}
			fakeUUID, _ := uuid.NewV4()
			fakeUUIDstr := fakeUUID.String()

			err := workflows.PublishCreateVMWorkflow(cpiConfig, fakeUUIDstr)
			Expect(err).ToNot(HaveOccurred())

			expectedWorkflow := workflows.GenerateCreateVMWorkflow(fakeUUIDstr)
			delete(expectedWorkflow.Options, "defaults")
			delete(expectedWorkflow.Options, "bootstrap-ubuntu")
			expectedReserveNodeTask := workflows.GenerateReserveNodeTask(fakeUUIDstr)
			delete(expectedReserveNodeTask.Options, "commands")
			expectedProvisionNodeTask := workflows.GenerateProvisionNodeTask(fakeUUIDstr)
			delete(expectedProvisionNodeTask.Options, "commands")

			workflowLibrary, err := onrackhttp.RetrieveWorkflows(cpiConfig)
			taskLibrary, err := onrackhttp.RetrieveTasks(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			for i := range taskLibrary {
				delete(taskLibrary[i].Options, "commands")
			}
			Expect(taskLibrary).To(ContainElement(expectedProvisionNodeTask))
			Expect(taskLibrary).To(ContainElement(expectedReserveNodeTask))


			for i := range workflowLibrary {
				delete(workflowLibrary[i].Options, "defaults")
				delete(workflowLibrary[i].Options, "bootstrap-ubuntu")
			}

			Expect(workflowLibrary).To(ContainElement(expectedWorkflow))
		})
	})

	Describe("InitiateCreateVMWorkflow", func(){
		It("successfully initiates the published workflow", func(){
			apiServer := os.Getenv("ON_RACK_API_URI")
			cpiConfig := config.Cpi{ApiServer: apiServer}
			fakeUUID, _ := uuid.NewV4()
			fakeUUIDstr := fakeUUID.String()
			nodeID := os.Getenv("ON_RACK_NODE_ID")

			defaults := onrackhttp.UploadAgentSettingsOptions{
				AgentSettingsFile:    nodeID,
				AgentSettingsPath:    onrackhttp.OnrackEnvPath,
				CID:                  "fake-vm-cid",
				RegistrySettingsFile: fmt.Sprintf("agent-%s", "fake-vm-cid"),
				RegistrySettingsPath: onrackhttp.OnrackRegistryPath,
				StemcellFile:         "12345678",
			}

			err := workflows.PublishCreateVMWorkflow(cpiConfig, fakeUUIDstr)
			Expect(err).To(BeNil())

			err = workflows.InitiateCreateVMWorkflow(cpiConfig, fakeUUIDstr, nodeID, defaults)
			Expect(err).To(BeNil())

			url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows", cpiConfig.ApiServer, nodeID)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			var nodeWorkflows []onrackhttp.NodeWorkflow
			json.Unmarshal(body, &nodeWorkflows)

			expectedNodeWorkflow := onrackhttp.NodeWorkflow{
				NodeID: 				nodeID,
				InjectableName:	fmt.Sprintf("Graph.CF.CreateReserveVM.%s", fakeUUIDstr),
				Status:					"valid",
			}

			Expect(nodeWorkflows).To(ContainElement(expectedNodeWorkflow))

			//todo: remove active workflow from node (dont worry about waiting for it to finish)
		})
	})
})
