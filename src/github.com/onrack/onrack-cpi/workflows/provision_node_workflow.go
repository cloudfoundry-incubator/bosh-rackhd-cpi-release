package workflows

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
)

func RunProvisionNodeWorkflow(c config.Cpi, nodeID string, workflowName string, options ProvisionNodeWorkflowOptions) error {
	req := onrackhttp.RunWorkflowRequestBody{
		Name:    workflowName,
		Options: map[string]interface{}{"defaults": options},
	}

	return onrackhttp.RunWorkflow(c, nodeID, req)
}

func PublishProvisionNodeWorkflow(cpiConfig config.Cpi, uuid string) (string, error) {
	tasks, workflow, err := generateProvisionNodeWorkflow(uuid)
	if err != nil {
		return "", err
	}

	for i := range tasks {
		err = onrackhttp.PublishTask(cpiConfig, tasks[i])
		if err != nil {
			return "", err
		}
	}

	w := provisionNodeWorkflow{}
	err = json.Unmarshal(workflow, &w)
	if err != nil {
		log.Printf("error umarshalling workflow: %s", err)
		return "", err
	}

	err = onrackhttp.PublishWorkflow(cpiConfig, workflow)
	if err != nil {
		return "", err
	}

	return w.Name, nil
}

func generateProvisionNodeWorkflow(uuid string) ([][]byte, []byte, error) {
	p := provisionNodeTask{}
	err := json.Unmarshal(provisionNodeTemplate, &p)
	if err != nil {
		log.Printf("error unmarshalling provision node task template: %s\n", err)
		return nil, nil, fmt.Errorf("error unmarshalling provision node task template: %s\n", err)
	}

	p.Name = fmt.Sprintf("%s.%s", p.Name, uuid)
	p.UnusedName = fmt.Sprintf("%s.%s", p.UnusedName, "UPLOADED_BY_ONRACK_CPI")

	pBytes, err := json.Marshal(p)
	if err != nil {
		log.Printf("error marshalling provision node task template: %s\n", err)
		return nil, nil, fmt.Errorf("error marshalling provision node task template: %s\n", err)
	}

	s := setNodeIDThenRebootTask{}
	err = json.Unmarshal(setNodeIDThenRebootTemplate, &s)
	if err != nil {
		log.Printf("error unmarshalling set node id task template: %s\n", err)
		return nil, nil, fmt.Errorf("error unmarshalling set node id task template: %s\n", err)
	}

	s.Name = fmt.Sprintf("%s.%s", s.Name, uuid)
	s.UnusedName = fmt.Sprintf("%s.%s", s.UnusedName, "UPLOADED_BY_ONRACK_CPI")

	sBytes, err := json.Marshal(s)
	if err != nil {
		log.Printf("error marshalling set node id task template: %s\n", err)
		return nil, nil, fmt.Errorf("error marshalling set node id task template: %s\n", err)
	}

	w := provisionNodeWorkflow{}
	err = json.Unmarshal(provisionNodeWorkflowTemplate, &w)
	if err != nil {
		log.Printf("error unmarshalling provision node workflow template: %s\n", err)
		return nil, nil, fmt.Errorf("error unmarshalling provision node workflow template: %s\n", err)
	}

	w.Name = fmt.Sprintf("%s.%s", w.Name, uuid)
	w.UnusedName = fmt.Sprintf("%s.%s", w.UnusedName, "UPLOADED_BY_ONRACK_CPI")
	for i := range w.Tasks {
		w.Tasks[i].TaskName = fmt.Sprintf("%s.%s", w.Tasks[i].TaskName, uuid)
	}

	wBytes, err := json.Marshal(w)
	if err != nil {
		log.Printf("error marshalling provision node workflow template: %s\n", err)
		return nil, nil, fmt.Errorf("error marshalling provision node workflow template: %s\n", err)
	}

	return [][]byte{pBytes, sBytes}, wBytes, nil
}

type ProvisionNodeWorkflowOptions struct {
	AgentSettingsFile    *string `json:"agentSettingsFile"`
	AgentSettingsPath    *string `json:"agentSettingsPath"`
	PublicKeyFile        *string `json:"publicKeyFile"`
	CID                  *string `json:"cid"`
	DownloadDir          string  `json:"downloadDir,omitempty"`
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

type provisionNodeWorkflow struct {
	*onrackhttp.WorkflowStub
	*provisionNodeWorkflowOptionsContainer
	Tasks []onrackhttp.WorkflowTask `json:"tasks"`
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
      "taskName": "Task.BOSH.Provision.Node"
    },
    {
      "label": "set-id-and-reboot",
      "taskName": "Task.BOSH.SetNodeId",
      "waitOn": {
        "provision-node": "succeeded"
      }
    }
  ]
}`)
