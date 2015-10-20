package workflows

import (
	"encoding/json"
	"fmt"

	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
)

var deprovisionNodeWorkflowTemplate = []byte(`{
  "friendlyName": "BOSH Deprovision Node",
  "injectableName": "Graph.BOSH.DeprovisionNode",
  "options": {},
  "tasks": [
    {
      "label": "set-boot-pxe",
      "taskName": "Task.Obm.Node.PxeBoot",
      "ignoreFailure": true
    },
    {
      "label": "reboot",
      "taskName": "Task.Obm.Node.Reboot",
      "waitOn": {
        "set-boot-pxe": "finished"
      }
    },
    {
      "label": "bootstrap-ubuntu",
      "taskName": "Task.Linux.Bootstrap.Ubuntu",
      "waitOn": {
        "reboot": "succeeded"
      }
    },
    {
      "label": "wipe-machine",
      "taskName": "Task.BOSH.DeprovisionNode",
      "waitOn": {
        "bootstrap-ubuntu": "succeeded"
      }
    },
    {
      "label": "shell-reboot",
      "taskName": "Task.ProcShellReboot",
      "waitOn": {
        "wipe-machine": "finished"
      }
    }
  ]
}`)

type deprovisionNodeWorkflow struct {
	*onrackapi.WorkflowStub
	*onrackapi.OptionContainer
	Tasks []onrackapi.WorkflowTask `json:"tasks"`
}

func RunDeprovisionNodeWorkflow(c config.Cpi, nodeID string, workflowName string) error {
	req := onrackapi.RunWorkflowRequestBody{
		Name:    workflowName,
		Options: map[string]interface{}{},
	}

	return onrackapi.RunWorkflow(onrackapi.WorkflowPoster, onrackapi.WorkflowFetcher, c, nodeID, req)
}

func PublishDeprovisionNodeWorkflow(c config.Cpi, uuid string) (string, error) {
	tasks, workflow, err := generateDeprovisionNodeWorkflow(uuid)
	if err != nil {
		return "", err
	}

	for i := range tasks {
		err = onrackapi.PublishTask(c, tasks[i])
		if err != nil {
			return "", err
		}
	}

	w := deprovisionNodeWorkflow{}
	err = json.Unmarshal(workflow, &w)
	if err != nil {
		return "", fmt.Errorf("error umarshalling workflow: %s", err)
	}

	err = onrackapi.PublishWorkflow(c, workflow)
	if err != nil {
		return "", err
	}

	return w.Name, nil
}

func generateDeprovisionNodeWorkflow(uuid string) ([][]byte, []byte, error) {
	deprovisionTask := deprovisionNodeTask{}
	err := json.Unmarshal(deprovisionNodeTaskTemplate, &deprovisionTask)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling Deprovision node task template: %s", err)
	}

	deprovisionTask.Name = fmt.Sprintf("%s.%s", deprovisionTask.Name, uuid)
	deprovisionTask.UnusedName = fmt.Sprintf("%s.%s", deprovisionTask.UnusedName, "UPLOADED_BY_ONRACK_CPI")

	deprovisionTaskBytes, err := json.Marshal(deprovisionTask)
	if err != nil {
		return nil, nil, fmt.Errorf("error Deprovision provision node task template: %s", err)
	}

	w := deprovisionNodeWorkflow{}
	err = json.Unmarshal(deprovisionNodeWorkflowTemplate, &w)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling Deprovision node workflow template: %s", err)
	}

	w.Name = fmt.Sprintf("%s.%s", w.Name, uuid)
	w.UnusedName = fmt.Sprintf("%s.%s", w.UnusedName, "UPLOADED_BY_ONRACK_CPI")
	w.Tasks[3].TaskName = fmt.Sprintf("%s.%s", w.Tasks[3].TaskName, uuid)

	wBytes, err := json.Marshal(w)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling Deprovision node workflow template: %s", err)
	}

	return [][]byte{deprovisionTaskBytes}, wBytes, nil
}
