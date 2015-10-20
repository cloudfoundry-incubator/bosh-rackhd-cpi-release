package workflows

import (
	"encoding/json"
	"fmt"

	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
)

var reserveNodeWorkflowTemplate = []byte(`{
  "friendlyName": "BOSH Reserve Node",
  "injectableName": "Graph.BOSH.ReserveNode",
  "options": {
    "defaults": {
      "uuid": null
    }
  },
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
      "label": "reserve-node",
      "taskName": "Task.BOSH.Reserve.Node",
      "waitOn": {
        "bootstrap-ubuntu": "succeeded"
      }
    }
  ]
}`)

type ReserveNodeWorkflowOptions struct {
	UUID *string `json:"uuid"`
}

type reserveNodeWorkflowOptionsContainer struct {
	Options reserveNodeWorkflowDefaultOptionsContainer `json:"options"`
}

type reserveNodeWorkflowDefaultOptionsContainer struct {
	Defaults ReserveNodeWorkflowOptions `json:"defaults"`
}

type reserveNodeWorkflow struct {
	*onrackapi.WorkflowStub
	*reserveNodeWorkflowOptionsContainer
	Tasks []onrackapi.WorkflowTask `json:"tasks"`
}

func RunReserveNodeWorkflow(c config.Cpi, nodeID string, workflowName string, options ReserveNodeWorkflowOptions) error {
	req := onrackapi.RunWorkflowRequestBody{
		Name:    workflowName,
		Options: map[string]interface{}{"defaults": options},
	}

	return onrackapi.RunWorkflow(onrackapi.WorkflowPoster, onrackapi.WorkflowFetcher, c, nodeID, req)
}

func PublishReserveNodeWorkflow(c config.Cpi, uuid string) (string, error) {
	tasks, workflow, err := generateReserveNodeWorkflow(uuid)
	if err != nil {
		return "", err
	}

	for i := range tasks {
		err = onrackapi.PublishTask(c, tasks[i])
		if err != nil {
			return "", err
		}
	}

	w := reserveNodeWorkflow{}
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

func generateReserveNodeWorkflow(uuid string) ([][]byte, []byte, error) {
	reserve := reserveNodeTask{}
	err := json.Unmarshal(reserveNodeTaskTemplate, &reserve)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling reserve node task template: %s", err)
	}

	reserve.Name = fmt.Sprintf("%s.%s", reserve.Name, uuid)
	reserve.UnusedName = fmt.Sprintf("%s.%s", reserve.UnusedName, "UPLOADED_BY_ONRACK_CPI")

	reserveBytes, err := json.Marshal(reserve)
	if err != nil {
		return nil, nil, fmt.Errorf("error reserve provision node task template: %s", err)
	}

	w := reserveNodeWorkflow{}
	err = json.Unmarshal(reserveNodeWorkflowTemplate, &w)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling reserve node workflow template: %s", err)
	}

	w.Name = fmt.Sprintf("%s.%s", w.Name, uuid)
	w.UnusedName = fmt.Sprintf("%s.%s", w.UnusedName, "UPLOADED_BY_ONRACK_CPI")
	w.Tasks[3].TaskName = fmt.Sprintf("%s.%s", w.Tasks[3].TaskName, uuid)

	wBytes, err := json.Marshal(w)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling reserve node workflow template: %s", err)
	}

	return [][]byte{reserveBytes}, wBytes, nil
}
