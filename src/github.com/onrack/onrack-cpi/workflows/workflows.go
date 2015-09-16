package workflows

import (
	"bytes"
	"fmt"
	// "errors"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"log"
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
	Options						Options									`json:"options"`
}

type Task struct {
	TaskName					string									`json:"taskName"`
	Label							string									`json:"label"`
	WaitOn						map[string]string				`json:"waitOn",omitempty`
	IgnoreFailure			bool										`json:"ignoreFailure",omitempty`
}

type Defaults struct {
	Cid 					string				`json:"cid"`
	DownloadDir 	string				`json:"downloadDir"`
	Env 					string				`json:"env"`
	File					string				`json:"file"`
	Path					string				`json:"path"`
}

type Options struct {
	BootstrapUbuntu			map[string]string		`json:"bootstrap-ubuntu"`
	Defaults						Defaults						`json:"defaults"`
}


func PublishCreateVMWorkflow(cpiConfig config.Cpi, uuid string) (err error) {
	createVMWorkflow := generateCreateVMWorkflowString(uuid)
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows", cpiConfig.ApiServer)
	body, err := json.Marshal(createVMWorkflow)
	if err != nil {
		log.Printf("error marshalling createVMWorkflow")
		return
	}

	request, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("error building http request")
		return
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("error sending PUT request to %s", cpiConfig.ApiServer)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Printf("error response code is %d: %s", resp.StatusCode, string(msg))
		return
	}

	return
}

func generateCreateVMWorkflowString(uuid string) (workflow Workflow) {
		workflow = Workflow{
			FriendlyName: "CF Create VM",
			InjectableName: fmt.Sprintf("Graph.CF.CreateVM.%s",uuid),
			Options: Options{
					BootstrapUbuntu: map[string]string{"overlayfs": "common/overlayfs_all_files.cf.cpio.gz"},
					Defaults: Defaults{
						Cid: "",
						Env: "",
						File: "",
						Path: "",
						DownloadDir: "/opt/downloads",
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
