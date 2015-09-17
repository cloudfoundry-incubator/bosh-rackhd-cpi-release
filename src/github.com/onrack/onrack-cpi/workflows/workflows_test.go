package workflows_test

import (
	"encoding/json"
	"fmt"
	"github.com/onrack/onrack-cpi/workflows"
	"github.com/onrack/onrack-cpi/config"
	"io/ioutil"
	"net/http"
	"github.com/nu7hatch/gouuid"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func expectedWorkflow(uuid string) (expectedWorkflow workflows.Workflow){
	expectedWorkflow = workflows.Workflow{
		FriendlyName: "CF Create VM",
		InjectableName: fmt.Sprintf("Graph.CF.CreateVM.%s", uuid),
		Options: workflows.Options{
				BootstrapUbuntu: map[string]string{"overlayfs": "common/overlayfs_all_files.cf.cpio.gz"},
				Defaults: workflows.Defaults{
					Cid: "",
					Env: "",
					File: "",
					Path: "",
					DownloadDir: "/opt/downloads",
				},
		},
		Tasks: []workflows.Task{
			workflows.Task{
				Label: "set-boot-pxe",
				TaskName: "Task.Obm.Node.PxeBoot",
				IgnoreFailure: true,
			},
			workflows.Task{
				Label: "reboot",
				TaskName: "Task.Obm.Node.Reboot",
				WaitOn: map[string]string{
					"set-boot-pxe": "finished",
				},
			},
			workflows.Task{
				Label: "bootstrap-ubuntu",
				TaskName: "Task.Linux.Bootstrap.Ubuntu",
				WaitOn: map[string]string{
					"reboot": "succeeded",
				},
			},
			workflows.Task{
				Label: "reserve-node",
				TaskName: "Task.Os.Reserve.CF.VM",
				WaitOn: map[string]string{
					"bootstrap-ubuntu": "succeeded",
				},
			},
			workflows.Task{
				Label: "download-stemcell",
				TaskName: "Task.Linux.DownloadFiles",
				WaitOn: map[string]string{
					"reserve-node": "succeeded",
				},
			},
			workflows.Task{
				Label: "stemcell-v2p-machine",
				TaskName: "Task.Os.Install.CF.Stemcell",
				WaitOn: map[string]string{
					"download-stemcell": "succeeded",
				},
			},
			workflows.Task{
				Label: "shell-reboot",
				TaskName: "Task.ProcShellReboot",
				WaitOn: map[string]string{
					"stemcell-v2p-machine": "finished",
				},
			},
		},
	}
	return
}

var _ = Describe("Workflows", func() {
	var apiServer string
	var cpiConfig config.Cpi
	var fakeUUID *uuid.UUID
	var nodeID string

	BeforeEach(func (){
			apiServer = os.Getenv("ON_RACK_API_URI")
			cpiConfig = config.Cpi{ApiServer: apiServer}
			fakeUUID, _ = uuid.NewV4()
			nodeID = os.Getenv("ON_RACK_NODE_ID")
	})

	Describe("PublishCreateVMWorkflow", func (){
		It("returns", func(){
			err := workflows.PublishCreateVMWorkflow(cpiConfig, fakeUUID.String())
			Expect(err).To(BeNil())

			url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/library", cpiConfig.ApiServer)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			var workflow_library []workflows.Workflow
			json.Unmarshal(body, &workflow_library)

			expectedWorkflow := expectedWorkflow(fakeUUID.String())

			Expect(workflow_library).To(ContainElement(expectedWorkflow))
		})
	})

	Describe("RunCreateVMWorkflow", func(){
		It("successfully runs the published workflow", func(){
			err := workflows.PublishCreateVMWorkflow(cpiConfig, fakeUUID.String())
			Expect(err).To(BeNil())
			url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows", cpiConfig.ApiServer, nodeID)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			var nodeWorkflows []workflows.NodeWorkflow
			json.Unmarshal(body, &nodeWorkflows)

			expectedNodeWorkflow := workflows.NodeWorkflow{
				NodeID: 				nodeID,
				InjectableName:	fmt.Sprintf("Graph.CF.CreateVM.%s", fakeUUID.String()),
				Status:					"succeeded",
			}
			Expect(nodeWorkflows).To(ContainElement(expectedNodeWorkflow))
		})
	})
})
