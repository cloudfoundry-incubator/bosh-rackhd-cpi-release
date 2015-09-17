package workflows

import (
	"bytes"
	"fmt"
	"encoding/json"
	"net/http"
	"io/ioutil"
	"log"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
)

//make sure they block until finished
//eg: poll workflow library, retry w/ timeout
//func PublishCreateVMWorkflow(config cpi.Config, uuid string) √
//func PublishDeleteVMWorkflow(config cpi.Config, uuid string)
//func UnpublishWorkflow(config cpi.Config, uuid string)
//func RunCreateVMWorkflow(config cpi.Config, nodeID string, uuid string)
//func RunDeleteVMWorkflow(config cpi.Config, nodeID string, uuid string)
//func KillActiveWorkflowsOnVM(config cpi.Config, nodeID string)

func PublishCreateVMWorkflow(cpiConfig config.Cpi, uuid string) (err error) {

	publishReserveNodeTask(cpiConfig, uuid)
	publishProvisonNodeTask(cpiConfig, uuid)

	createVMWorkflow := GenerateCreateVMWorkflow(uuid)
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

func GenerateCreateVMWorkflow(uuid string) (workflow onrackhttp.Workflow) {
		workflow = onrackhttp.Workflow{
			FriendlyName: "CF CreateReserve VM",
			InjectableName: fmt.Sprintf("Graph.CF.CreateReserveVM.%s",uuid),
			Options: onrackhttp.Options{
					BootstrapUbuntu: map[string]string{"overlayfs": "common/overlayfs_all_files.cf.cpio.gz"},
					Defaults: onrackhttp.Defaults{
						AgentSettingsFile: "",
						AgentSettingsPath: "",
						Cid: "",
						DownloadDir: "/opt/downloads",
						RegistrySettingsFile: "",
						RegistrySettingsPath: "",
						StemcellFile: "",
					},
			},
			Tasks: []onrackhttp.WorkflowTask{
				onrackhttp.WorkflowTask{
					Label: "set-boot-pxe",
					TaskName: "Task.Obm.Node.PxeBoot",
					IgnoreFailure: true,
				},
				onrackhttp.WorkflowTask{
					Label: "reboot",
					TaskName: "Task.Obm.Node.Reboot",
					WaitOn: map[string]string{
						"set-boot-pxe": "finished",
					},
				},
				onrackhttp.WorkflowTask{
					Label: "bootstrap-ubuntu",
					TaskName: "Task.Linux.Bootstrap.Ubuntu",
					WaitOn: map[string]string{
						"reboot": "succeeded",
					},
				},
				onrackhttp.WorkflowTask{
					Label: "reserve-node",
					TaskName: fmt.Sprintf("Task.Os.Reserve.CF.VM.%s", uuid),
					WaitOn: map[string]string{
						"bootstrap-ubuntu": "succeeded",
					},
				},
				onrackhttp.WorkflowTask{
					Label: "provision-node",
					TaskName: fmt.Sprintf("Task.Os.Install.CF.Stemcell.%s", uuid),
					WaitOn: map[string]string{
						"reserve-node": "succeeded",
					},
				},
				onrackhttp.WorkflowTask{
					Label: "shell-reboot",
					TaskName: "Task.ProcShellReboot",
					WaitOn: map[string]string{
						"provision-node": "finished",
					},
				},
			},
	}
	return
}

func publishReserveNodeTask(cpiConfig config.Cpi, uuid string) (err error) {
	task := GenerateReserveNodeTask(uuid)
	onrackhttp.PublishTask(cpiConfig, task)
	return
}

func GenerateReserveNodeTask(uuid string) (task onrackhttp.Task) {
	task = onrackhttp.Task{
		FriendlyName: "Reserve Node",
		ImplementsTask: "Task.Base.Linux.Commands",
		InjectableName: fmt.Sprintf("Task.Os.Reserve.CF.VM.%s", uuid),
		Options: map[string]interface{} {
			"cid": nil,
			"commands": []string{
				"curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"reserved\": \"{{ options.cid }}\" }'",
			},
		},
		Properties: map[string]interface{}{},
	}
	return
}

func publishProvisonNodeTask(cpiConfig config.Cpi, uuid string) (err error) {
 	task := GenerateProvisionNodeTask(uuid)
	onrackhttp.PublishTask(cpiConfig, task)
	return
}

func GenerateProvisionNodeTask(uuid string) (task onrackhttp.Task) {
	task = onrackhttp.Task{
		FriendlyName: "Provision Node",
		ImplementsTask: "Task.Base.Linux.Commands",
		InjectableName: fmt.Sprintf("Task.Os.Install.CF.Stemcell.%s", uuid),
		Options: map[string]interface{} {
			"agentSettingsFile": "",
			"agentSettingsMd5Uri": "{{ api.files }}/md5/{{ options.agentSettingsFile }}/latest",
			"agentSettingsPath": "",
			"agentSettingsUri": "{{ api.files }}/{{ options.agentSettingsFile }}/latest",
			"commands": []string{
				"curl --retry 3 {{ options.stemcellUri }} -o {{ options.downloadDir }}/{{ options.stemcellFile }}",
        "test `curl {{ options.stemcellFileMd5Uri }}` = \\\"`md5sum {{ options.downloadDir }}/{{ options.stemcellFile }}| awk '{print $1}'`\\\"",
        "curl --retry 3 {{ options.agentSettingsUri }} -o {{ options.downloadDir }}/{{ options.agentSettingsFile }}",
        "test `curl {{ options.agentSettingsMd5Uri }}` = \\\"`md5sum {{ options.downloadDir }}/{{ options.agentSettingsFile }}| awk '{print $1}'`\\\"",
        "curl --retry 3 {{ options.registrySettingsUri }} -o {{ options.downloadDir }}/{{ options.registrySettingsFile }}",
        "test `curl {{ options.registrySettingsMd5Uri }}` = \\\"`md5sum {{ options.downloadDir }}/{{ options.registrySettingsFile }}| awk '{print $1}'`\\\"",
        "sudo qemu-img convert {{ options.downloadDir }}/{{ options.stemcellFile }} -O raw {{ options.device }}",
        "sudo mount {{ options.device }}1 /mnt",
        "sudo cp {{ options.downloadDir }}/{{ options.agentSettingsFile }} /mnt/{{ options.agentSettingsPath }}",
        "sudo cp {{ options.downloadDir }}/{{ options.registrySettingsFile }} /mnt/{{ options.registrySettingsPath }}",
        "sudo sync",
			},
			"device": "/dev/sda",
      "downloadDir": "/opt/downloads",
      "registrySettingsFile": "",
      "registrySettingsMd5Uri": "{{ api.files }}/md5/{{ options.registrySettingsFile }}/latest",
      "registrySettingsPath": "",
      "registrySettingsUri": "{{ api.files }}/{{ options.registrySettingsFile }}/latest",
      "stemcellFile": "",
      "stemcellFileMd5Uri": "{{ api.files }}/md5/{{ options.stemcellFile }}/latest",
      "stemcellUri": "{{ api.files }}/{{ options.stemcellFile }}/latest",
		},
		Properties: map[string]interface{}{},
	}
	return
}
