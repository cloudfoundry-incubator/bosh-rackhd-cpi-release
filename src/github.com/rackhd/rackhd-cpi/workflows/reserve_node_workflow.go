package workflows

import (
	"encoding/json"
	"fmt"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

type reserveNodeWorkflowOptions struct {
	OBMServiceName *string `json:"obmServiceName"`
}

type reserveNodeWorkflowDefaultOptionsContainer struct {
	Defaults reserveNodeWorkflowOptions `json:"defaults"`
}

type reserveNodeWorkflowOptionsContainer struct {
	Options reserveNodeWorkflowDefaultOptionsContainer `json:"options"`
}

type reserveNodeWorkflow struct {
	*models.Graph
	*reserveNodeWorkflowOptionsContainer
	Tasks []models.WorkflowTask `json:"tasks"`
}

func RunReserveNodeWorkflow(c config.Cpi, nodeID string, workflowName string) error {
	options, err := buildReserveNodeWorkflowOptions(c, nodeID)
	if err != nil {
		return err
	}

	req := models.RunWorkflowRequestBody{
		Name:    workflowName,
		Options: map[string]interface{}{"defaults": options},
	}

	return rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, c, nodeID, req)
}

func PublishReserveNodeWorkflow(c config.Cpi) (string, error) {
	tasks, workflow, err := generateReserveNodeWorkflow(c.RequestID)
	if err != nil {
		return "", err
	}

	for i := range tasks {
		err = rackhdapi.PublishTask(c, tasks[i])
		if err != nil {
			return "", err
		}
	}

	w := reserveNodeWorkflow{}
	err = json.Unmarshal(workflow, &w)
	if err != nil {
		return "", fmt.Errorf("error umarshalling workflow: %s", err)
	}

	err = rackhdapi.PublishGraph(c, workflow)
	if err != nil {
		return "", err
	}

	return w.Name, nil
}

func generateReserveNodeWorkflow(uuid string) ([][]byte, []byte, error) {
	reserveNodeTaskBytes, err := helpers.ReadFile("../templates/reserve_node_task.json")
	if err != nil {
		return nil, nil, err
	}

	reserve := reserveNodeTask{}
	err = json.Unmarshal(reserveNodeTaskBytes, &reserve)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling reserve node task template: %s", err)
	}

	reserve.Name = fmt.Sprintf("%s.%s", reserve.Name, uuid)
	reserve.UnusedName = fmt.Sprintf("%s.%s", reserve.UnusedName, "UPLOADED_BY_RACKHD_CPI")

	reserveBytes, err := json.Marshal(reserve)
	if err != nil {
		return nil, nil, fmt.Errorf("error reserve provision node task template: %s", err)
	}

	reserveNodeWorkflowBytes, err := helpers.ReadFile("../templates/reserve_node_workflow.json")
	if err != nil {
		return nil, nil, err
	}

	w := reserveNodeWorkflow{}
	err = json.Unmarshal(reserveNodeWorkflowBytes, &w)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling reserve node workflow template: %s", err)
	}

	w.Name = fmt.Sprintf("%s.%s", w.Name, uuid)
	w.UnusedName = fmt.Sprintf("%s.%s", w.UnusedName, "UPLOADED_BY_RACKHD_CPI")
	w.Tasks[3].TaskName = fmt.Sprintf("%s.%s", w.Tasks[3].TaskName, uuid)

	wBytes, err := json.Marshal(w)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling reserve node workflow template: %s", err)
	}

	return [][]byte{reserveBytes}, wBytes, nil
}

func buildReserveNodeWorkflowOptions(c config.Cpi, nodeID string) (reserveNodeWorkflowOptions, error) {
	options := reserveNodeWorkflowOptions{}

	obmServiceName, err := rackhdapi.GetOBMServiceName(c, nodeID)
	if err != nil {
		return reserveNodeWorkflowOptions{}, err
	}
	options.OBMServiceName = &obmServiceName

	return options, nil
}
