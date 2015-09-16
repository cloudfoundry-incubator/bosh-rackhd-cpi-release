package workflows

import (
	"fmt"
	// "errors"
//	"encoding/json"
	"github.com/onrack/onrack-cpi/config"
)

//make sure they block until finished
//eg: poll workflow library, retry w/ timeout
//func PublishCreateVMWorkflow(config cpi.Config, uuid string)
//func PublishDeleteVMWorkflow(config cpi.Config, uuid string)
//func UnpublishWorkflow(config cpi.Config, uuid string)
//func RunCreateVMWorkflow(config cpi.Config, nodeID string, uuid string)
//func RunDeleteVMWorkflow(config cpi.Config, nodeID string, uuid string)
//func KillActiveWorkflowsOnVM(config cpi.Config, nodeID string)

type Workflow struct {
	FriendlyName			string 									`json:"friendlyName"`
	InjectableName 		string 									`json:"injectableName"`
	Tasks 						[]Task 									`json:"tasks"`
	Options						map[string]interface{}	`json:"options"`
}

type Task struct {
	TaskName					string									`json:"taskName"`
	Label							string									`json:"label"`
	WaitOn						map[string]string				`json:"waitOn",omitempty`
	IgnoreFailure			bool										`json:"ignoreFailure",omitempty`
}


func PublishCreateVMWorkflow(cpiConfig config.Cpi, uuid string) (err error) {
	createVMWorkflow := generateCreateVMWorkflowString(uuid)
	// http.GetClient.POST(url, createVMWorkflow)
	fmt.Printf("\n%s\n", createVMWorkflow)
	return
}

func generateCreateVMWorkflowString(uuid string) (workflow Workflow) {
		workflow = Workflow{
			FriendlyName: "CF Create VM",
			InjectableName: fmt.Sprintf("Graph.CF.CreateVM.%s",uuid),
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
			Tasks: []Task{
				Task{
					Label: "set-boot-pxe",
					TaskName: "Task.Obm.Node.PxeBoot",
					IgnoreFailure: true,
				},
				Task{
					Label: "reboot",
					TaskName: "Task.Obm.Node.Reboot",
					WaitOn: map[string]string{
						"set-boot-pxe": "finished",
					},
				},
				Task{
					Label: "bootstrap-ubuntu",
					TaskName: "Task.Linux.Bootstrap.Ubuntu",
					WaitOn: map[string]string{
						"reboot": "succeeded",
					},
				},
				Task{
					Label: "reserve-node",
					TaskName: "Task.Os.Reserve.CF.VM",
					WaitOn: map[string]string{
						"bootstrap-ubuntu": "succeeded",
					},
				},
				Task{
					Label: "download-stemcell",
					TaskName: "Task.Linux.DownloadFiles",
					WaitOn: map[string]string{
						"reserve-node": "succeeded",
					},
				},
				Task{
					Label: "stemcell-v2p-machine",
					TaskName: "Task.Os.Install.CF.Stemcell",
					WaitOn: map[string]string{
						"download-stemcell": "succeeded",
					},
				},
				Task{
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
