package workflows

import (
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
)

//make sure they block until finished
//eg: poll workflow library, retry w/ timeout
//func PublishCreateVMWorkflow(config cpi.Config, uuid string)
//func PublishDeleteVMWorkflow(config cpi.Config, uuid string)
//func UnpublishWorkflow(config cpi.Config, uuid string)
//func RunCreateVMWorkflow(config cpi.Config, nodeID string, uuid string)
//func RunDeleteVMWorkflow(config cpi.Config, nodeID string, uuid string)
//func KillActiveWorkflowsOnVM(config cpi.Config, nodeID string)

func PublishProvisionNodeWorkflow(cpiConfig config.Cpi, uuid string) error {

	return nil
}

func GenerateProvisionNodeWorkflow(uuid string) onrackhttp.Workflow {

	return onrackhttp.Workflow{}
}

type ProvisionNodeWorkflowOptions struct {
	AgentSettingsFile    *string `json:"agentSettingsFile"`
	AgentSettingsPath    *string `json:"agentSettingsPath"`
	PublicKeyFile        *string `json:"publicKeyFile"`
	CID                  *string `json:"cid"`
	DownloadDir          string  `json:"downloadDir"`
	RegistrySettingsFile *string `json:"registrySettingsFile"`
	RegistrySettingsPath *string `json:"registrySettingsPath"`
	StemcellFile         *string `json:"stemcellFile"`
}

type provisionNodeWorkflowOptionsContainer struct {
	Options provisionNodeWorkflowDefaultOptionsContainer `json:"options"`
}

type provisionNodeWorkflowDefaultOptionsContainer struct {
	Defaults ProvisionNodeWorkflowOptions `json:"defaults"`
}

type provisionNodeTasksContainer struct {
	TaskList []onrackhttp.WorkflowTask `json:"tasks"`
}

type ProvisionNodeWorkflow struct {
	*onrackhttp.WorkflowStub
	*onrackhttp.PropertyContainer
	*provisionNodeWorkflowOptionsContainer
	*provisionNodeTasksContainer
}

var provisionNodeWorkflowTemplate = []byte(`{
  "friendlyName": "BOSH Provision Node",
  "injectableName": "Graph.BOSH.ProvisionNode",
  "options": {
    "defaults": {
      "agentSettingsFile": null,
      "agentSettingsPath": null,
      "publicKeyFile": null,
      "cid": null,
      "downloadDir": "/opt/downloads",
      "registrySettingsFile": null,
      "registrySettingsPath": null,
      "stemcellFile": null
    }
  },
  "tasks": [
    {
      "label": "provision-node",
      "taskName": "Task.Provision.BOSH.Node"
    },
    {
      "label": "set-id-and-reboot",
      "taskName": "Task.SetId.VM",
      "waitOn": {
        "provision-node": "succeeded"
      }
    }
  ]
}`)
