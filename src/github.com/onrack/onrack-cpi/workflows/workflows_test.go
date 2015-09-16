package workflows_test

import (
	"encoding/json"
	"fmt"
	"github.com/onrack/onrack-cpi/workflows"
	"github.com/onrack/onrack-cpi/config"
	"io/ioutil"
	"net/http"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workflows", func() {
	Describe("PublishCreateVMWorkflow", func (){
		It("returns", func(){
			api_server := "192.168.252.131"
			config := config.Cpi{ApiServer: api_server}
			fakeUuid, _ := uuid.NewV4()
			err := workflows.PublishCreateVMWorkflow(config, fakeUuid.String())
			Expect(err).To(BeNil())

			url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/library", config.ApiServer)
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			var workflow_library []workflows.Workflow
			json.Unmarshal(body, &workflow_library)

			expectedWorkflow := workflows.Workflow{
				FriendlyName: "CF Create VM",
				InjectableName: fmt.Sprintf("Graph.CF.CreateVM.%s", fakeUuid.String()),
				Options: map[string]interface{}{
					"defaults": map[string]interface{}{
						"cid": nil,
						"downloadDir": "/opt/downloads",
						"env": nil,
		        "file": nil,
		        "path": nil,
					},
					"bootstrap-ubuntu": map[string]string{
						"overlayfs": "common/overlayfs_all_files.cf.cpio.gz",
					},
				},
				Tasks: []workflows.Task{
					workflows.Task{
						Label: "set-boot-pxe",
						IgnoreFailure: true,
						TaskName: "Task.Obm.Node.PxeBoot",
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



			Expect(workflow_library).To(ContainElement(expectedWorkflow))
		})
	})
})
