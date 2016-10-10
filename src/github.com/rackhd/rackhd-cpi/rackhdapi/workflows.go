package rackhdapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
)

const (
	OBMSettingIPMIServiceName = "ipmi-obm-service"
	OBMSettingAMTServiceName  = "amt-obm-service"
)

const (
	workflowSuccessfulStatus = "succeeded"
	workflowFailedStatus     = "failed"
	workflowCancelledStatus  = "cancelled"
	workflowRunningStatus    = "running"
	workflowPendingStatus    = "pending"
)

const (
	RackHDReserveVMGraphName = "Graph.CF.ReserveVM"
	RackHDCreateVMGraphName  = "Graph.BOSH.ProvisionNode"
	RackHDDeleteVMGraphName  = "Graph.CF.DeleteVM"
	RackHDEnvPath            = "/var/vcap/bosh/agent-bootstrap-env.json"
	DefaultUnusedName        = "UPLOADED_BY_RACKHD_CPI"
)

type NodeWorkflow struct {
	NodeID         string `json:"node"`
	InjectableName string `json:"injectableName"`
	Status         string `json:"_status"`
}

type Workflow struct {
	Name       string                 `json:"injectableName"`
	UnusedName string                 `json:"friendlyName"`
	Tasks      []WorkflowTask         `json:"tasks"`
	Options    map[string]interface{} `json:"options"`
}

type Graph struct {
	Name       string                 `json:"injectableName"`
	UnusedName string                 `json:"friendlyName"`
	Options    map[string]interface{} `json:"options"`
	Tasks      []WorkflowTask         `json:"tasks"`
}

type WorkflowTask struct {
	TaskName string             `json:"taskName"`
	Label    string             `json:"label"`
	WaitOn   *map[string]string `json:"waitOn,omitempty"`
}

type WorkflowResponse struct {
	Name       string `json:"injectableName"`
	Status     string `json:"status"`
	InstanceID string `json:"instanceId"`
}

type PropertyContainer struct {
	Properties interface{} `json:"properties"`
}

type OptionContainer struct {
	Options interface{} `json:"options"`
}

type RunWorkflowRequestBody struct {
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}

type workflowFetcherFunc func(config.Cpi, string) (WorkflowResponse, error)

type workflowPosterFunc func(config.Cpi, string, RunWorkflowRequestBody) (WorkflowResponse, error)

func PublishWorkflow(c config.Cpi, graphBytes []byte) error {
	url := fmt.Sprintf("%s/api/2.0/workflows/graphs", c.ApiServer)
	fmt.Printf("\nrequest body: %+v\n", string(graphBytes))
	log.Debug(fmt.Sprintf("workflow to publish: %s", string(graphBytes)))
	request, err := http.NewRequest("PUT", url, bytes.NewReader(graphBytes))
	request.Close = true

	if err != nil {
		return fmt.Errorf("error building http request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)

	fmt.Printf("\n\n\nreq: %+v\n body: %+v", request, string(graphBytes))
	if err != nil {
		return fmt.Errorf("error sending publishing workflow to %s", url)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("\npublish body: %+v\n", string(b))

	if resp.StatusCode != 201 {
		return fmt.Errorf("error publishing workflow; response status code: %s,\nresponse body: %+v", resp.Status, resp)
	}

	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}

	graph := Graph{}
	err = json.Unmarshal(graphBytes, &graph)
	if err != nil {
		return fmt.Errorf("error unmarshalling graph: %s", err)
	}
	log.Debug("workflow received after publishing: %s", string(graphBytes))

	publishedWorkflowsBytes, err := RetrieveWorkflows(c)
	if err != nil {
		return err
	}

	publishedWorkflows := []Graph{}
	err = json.Unmarshal(publishedWorkflowsBytes, &publishedWorkflows)
	if err != nil {
		return fmt.Errorf("error unmarshalling published workflows: %s", err)
	}

	var uploadedWorkflow *Graph
	for i := range publishedWorkflows {
		if publishedWorkflows[i].Name == graph.Name {
			uploadedWorkflow = &publishedWorkflows[i]
		}
	}

	if uploadedWorkflow == nil {
		return errors.New("workflow was not successfully uploaded to server")
	}

	return nil
}

func RetrieveWorkflows(c config.Cpi) ([]byte, error) {
	url := fmt.Sprintf("%s/api/2.0/workflows/graphs", c.ApiServer)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error making request: %s", err)
	}
	request.Close = true

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed retrieving workflows with status: %s, message: %s", resp.Status, string(body))
	}

	return body, nil
}

func WorkflowFetcher(c config.Cpi, graphName string) (WorkflowResponse, error) {
	url := fmt.Sprintf("%s/api/2.0/workflows/%s", c.ApiServer, graphName)
	resp, err := http.Get(url)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("Error requesting workflow on node at url: %s, msg: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return WorkflowResponse{}, fmt.Errorf("Failed retrieving workflow at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}

	body, err := ioutil.ReadAll(resp.Body)
	var workflow WorkflowResponse
	err = json.Unmarshal(body, &workflow)
	if err != nil {
		fmt.Printf("workflow fetcher: %+v", string(body))
		return WorkflowResponse{}, fmt.Errorf("Error unmarshalling workflow: %s", err)
	}

	return workflow, nil
}

func WorkflowPoster(c config.Cpi, nodeID string, req RunWorkflowRequestBody) (WorkflowResponse, error) {
	url := fmt.Sprintf("%s/api/2.0/nodes/%s/workflows", c.ApiServer, nodeID)
	body, err := json.Marshal(req)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error marshalling workflow request body, %s", err)
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error building http request to run workflow, %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	log.Debug("Posting workflow...")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error running workflow at url %s", url)
	}
	defer resp.Body.Close()
	bo, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("\n\npost body: %+v\n\n", string(bo))

	if resp.StatusCode != 201 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return WorkflowResponse{}, fmt.Errorf("Failed running workflow at url: %s with status: %s, message: %s, body: %s", url, resp.Status, string(msg), string(body))
	}

	wfRespBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error reading workflow response body %s", err)
	}

	workflowResp := WorkflowResponse{}
	fmt.Printf("workflow response in poster: %+v\n\n", string(wfRespBytes))
	err = json.Unmarshal(wfRespBytes, &workflowResp)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error unmarshalling /common/node/workflows response %s", err)
	}

	log.Debug("Workflow post successful")
	return workflowResp, nil
}

func RunWorkflow(poster workflowPosterFunc, fetcher workflowFetcherFunc, c config.Cpi, nodeID string, req RunWorkflowRequestBody) error {
	postedWorkflow, err := poster(c, nodeID, req)
	if err != nil {
		return fmt.Errorf("Failed to post workflow: %s", err)
	}

	timeoutChan := time.NewTimer(time.Second * c.RunWorkflowTimeoutSeconds).C
	retryChan := time.NewTicker(time.Second * 3).C

	for {
		select {
		case <-timeoutChan:
			err := KillActiveWorkflow(c, nodeID)
			if err != nil {
				return fmt.Errorf("Could not abort timed out workflow on node: %s", nodeID)
			}
			return fmt.Errorf("Timed out running workflow: %s on node: %s", req.Name, nodeID)

		case <-retryChan:
			wr, err := fetcher(c, postedWorkflow.InstanceID)
			if err != nil {
				return fmt.Errorf("Unable to fetch workflow status: %s", err)
			}

			switch wr.Status {
			case workflowRunningStatus:
				log.Info(fmt.Sprintf("workflow: %s is running against node: %s", req.Name, nodeID))
				continue
			case workflowSuccessfulStatus:
				log.Info(fmt.Sprintf("workflow: %s completed successfully against node: %s", req.Name, nodeID))
				return nil
			case workflowFailedStatus:
				return fmt.Errorf("workflow: %s failed against node: %s", req.Name, nodeID)
			case workflowCancelledStatus:
				log.Info(fmt.Sprintf("workflow: %s was cancelled against node: %s", req.Name, nodeID))
				return nil
			case workflowPendingStatus:
				log.Info(fmt.Sprintf("workflow: %s is pending on node: %s", req.Name, nodeID))
				continue
			default:
				return fmt.Errorf("workflow: %s has unexpected status '%s' on node: %s", req.Name, wr.Status, nodeID)
			}
		}
	}
}

func KillActiveWorkflow(c config.Cpi, nodeID string) error {
	url := fmt.Sprintf("%s/api/1.1/nodes/%s/workflows/active", c.ApiServer, nodeID)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error: %s building http request to delete active workflows against node: %s", err, nodeID)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error: %s deleting active workflows on node: %s", err, nodeID)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		msg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error: %s reading response body from abort workflows request against node: %s", err, nodeID)
		}

		return fmt.Errorf("Failed deleting active workflows against node: %s with status: %s, message: %s", nodeID, resp.Status, string(msg))
	}

	return nil
}

func GetActiveWorkflows(c config.Cpi, nodeID string) (WorkflowResponse, error) {
	var workflows WorkflowResponse

	url := fmt.Sprintf("%s/api/1.1/nodes/%s/workflows/active", c.ApiServer, nodeID)
	resp, err := http.Get(url)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("Error requesting active workflows on node at url: %s, msg: %s", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return WorkflowResponse{}, nil
	}

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return WorkflowResponse{}, fmt.Errorf("Failed retrieving active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &workflows)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("Error unmarshalling active workflows: %s %s", err, string(body))
	}

	return workflows, nil
}
