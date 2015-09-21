package workflows

import (
	"fmt"
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
	err = publishReserveNodeTask(cpiConfig, uuid)
	if err != nil {
		log.Printf("error publishing reserve node task to %s", cpiConfig.ApiServer)
		return
	}
	err = publishProvisonNodeTask(cpiConfig, uuid)
	if err != nil {
		log.Printf("error publishing provision node task to %s", cpiConfig.ApiServer)
		return
	}

	createVMWorkflow := GenerateCreateVMWorkflow(uuid)
	err = onrackhttp.PublishWorkflow(cpiConfig, createVMWorkflow)
	return
}

func GenerateCreateVMWorkflow(uuid string) (workflow onrackhttp.Workflow) {
	workflow = onrackhttp.Workflow{
		FriendlyName:   "CF CreateReserve VM",
		InjectableName: fmt.Sprintf("Graph.CF.CreateReserveVM.%s", uuid),
		Options: map[string]interface{}{
			"bootstrap-ubuntu": map[string]string{"overlayfs": "common/overlayfs_all_files.cpio.gz"},
			"defaults": map[string]interface{}{
				"agentSettingsFile":    nil,
				"agentSettingsPath":    nil,
				"cid":                  nil,
				"downloadDir":          "/opt/downloads",
				"registrySettingsFile": nil,
				"registrySettingsPath": nil,
				"stemcellFile":         nil,
			},
		},
		Tasks: []onrackhttp.WorkflowTask{
			onrackhttp.WorkflowTask{
				Label:         "set-boot-pxe",
				TaskName:      "Task.Obm.Node.PxeBoot",
				IgnoreFailure: true,
			},
			onrackhttp.WorkflowTask{
				Label:    "reboot",
				TaskName: "Task.Obm.Node.Reboot",
				WaitOn: map[string]string{
					"set-boot-pxe": "finished",
				},
			},
			onrackhttp.WorkflowTask{
				Label:    "bootstrap-ubuntu",
				TaskName: "Task.Linux.Bootstrap.Ubuntu",
				WaitOn: map[string]string{
					"reboot": "succeeded",
				},
			},
			onrackhttp.WorkflowTask{
				Label:    "reserve-node",
				TaskName: fmt.Sprintf("Task.Os.Reserve.CF.VM.%s", uuid),
				WaitOn: map[string]string{
					"bootstrap-ubuntu": "succeeded",
				},
			},
			onrackhttp.WorkflowTask{
				Label:    "provision-node",
				TaskName: fmt.Sprintf("Task.Os.Install.CF.Stemcell.%s", uuid),
				WaitOn: map[string]string{
					"reserve-node": "succeeded",
				},
			},
			onrackhttp.WorkflowTask{
				Label:    "shell-reboot",
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
	err = onrackhttp.PublishTask(cpiConfig, task)
	return
}

func GenerateReserveNodeTask(uuid string) (task onrackhttp.Task) {
	task = onrackhttp.Task{
		FriendlyName:   "Reserve Node",
		ImplementsTask: "Task.Base.Linux.Commands",
		InjectableName: fmt.Sprintf("Task.Os.Reserve.CF.VM.%s", uuid),
		Options: map[string]interface{}{
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
		FriendlyName:   "Provision Node",
		ImplementsTask: "Task.Base.Linux.Commands",
		InjectableName: fmt.Sprintf("Task.Os.Install.CF.Stemcell.%s", uuid),
		Options: map[string]interface{}{
			"agentSettingsFile":   nil,
			"agentSettingsMd5Uri": "{{ api.files }}/md5/{{ options.agentSettingsFile }}/latest",
			"agentSettingsPath":   nil,
			"agentSettingsUri":    "{{ api.files }}/{{ options.agentSettingsFile }}/latest",
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
			"device":                 "/dev/sda",
			"downloadDir":            "/opt/downloads",
			"registrySettingsFile":   nil,
			"registrySettingsMd5Uri": "{{ api.files }}/md5/{{ options.registrySettingsFile }}/latest",
			"registrySettingsPath":   nil,
			"registrySettingsUri":    "{{ api.files }}/{{ options.registrySettingsFile }}/latest",
			"stemcellFile":           nil,
			"stemcellFileMd5Uri":     "{{ api.files }}/md5/{{ options.stemcellFile }}/latest",
			"stemcellUri":            "{{ api.files }}/{{ options.stemcellFile }}/latest",
		},
		Properties: map[string]interface{}{},
	}
	return
}

func InitiateCreateVMWorkflow(cpiConfig config.Cpi, uuid string, nodeID string, options onrackhttp.UploadAgentSettingsOptions) (err error) {
	var body onrackhttp.RunWorkflowRequestBody
	if options.DownloadDir != "" {
		body = onrackhttp.RunWorkflowRequestBody{
			Name: fmt.Sprintf("Graph.CF.CreateReserveVM.%s", uuid),
			Options: map[string]interface{}{
				"defaults": map[string]interface{}{
					"agentSettingsFile":    options.AgentSettingsFile,
					"agentSettingsPath":    options.AgentSettingsPath,
					"cid":                  options.CID,
					"downloadDir":          options.DownloadDir,
					"registrySettingsFile": options.RegistrySettingsFile,
					"registrySettingsPath": options.RegistrySettingsPath,
					"stemcellFile":         options.StemcellFile,
				},
			},
		}
	} else {
		body = onrackhttp.RunWorkflowRequestBody{
			Name: fmt.Sprintf("Graph.CF.CreateReserveVM.%s", uuid),
			Options: map[string]interface{}{
				"defaults": map[string]interface{}{
					"agentSettingsFile": options.AgentSettingsFile,
					"agentSettingsPath": options.AgentSettingsPath,
					"cid":               options.CID,
					"registrySettingsFile": options.RegistrySettingsFile,
					"registrySettingsPath": options.RegistrySettingsPath,
					"stemcellFile":         options.StemcellFile,
				},
			},
		}
	}
	err = onrackhttp.InitiateWorkflow(cpiConfig, nodeID, body)
	return
}
