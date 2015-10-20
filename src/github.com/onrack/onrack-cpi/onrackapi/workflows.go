package onrackapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/onrack/onrack-cpi/config"
)

const (
	workflowValidStatus      = "valid"
	worfklowSuccessfulStatus = "succeeded"
	workflowFailedStatus     = "failed"
	workflowCancelledStatus  = "cancelled"
)

const (
	OnrackReserveVMGraphName = "Graph.CF.ReserveVM"
	OnrackCreateVMGraphName  = "Graph.BOSH.ProvisionNode"
	OnrackDeleteVMGraphName  = "Graph.CF.DeleteVM"
	OnrackEnvPath            = "/var/vcap/bosh/agent-bootstrap-env.json"
	DefaultUnusedName        = "UPLOADED_BY_ONRACK_CPI"
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

type WorkflowStub struct {
	Name       string         `json:"injectableName"`
	UnusedName string         `json:"friendlyName"`
	Tasks      []WorkflowTask `json:"tasks"`
}

type WorkflowResponse struct {
	Name         string                  `json:"injectableName"`
	Tasks        map[string]TaskResponse `json:"tasks"`
	Status       string                  `json:"_status"`
	ID           string                  `json:"id"`
	PendingTasks []interface{}           `json:"pendingTasks"`
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

type workflowFetcherFunc func(config.Cpi, string, string) (WorkflowResponse, error)

type workflowPosterFunc func(config.Cpi, string, RunWorkflowRequestBody) (WorkflowResponse, error)

func PublishWorkflow(c config.Cpi, workflowBytes []byte) error {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows", c.ApiServer)

	request, err := http.NewRequest("PUT", url, bytes.NewReader(workflowBytes))
	request.Close = true

	if err != nil {
		return fmt.Errorf("error building http request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("error sending publishing workflow to %s", url)
	}
	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed publishing workflows with status: %s, body: %s", resp.Status, string(msg))
	}

	workflowStub := WorkflowStub{}
	err = json.Unmarshal(workflowBytes, &workflowStub)
	if err != nil {
		return fmt.Errorf("error unmarshalling workflow: %s", err)
	}

	publishedWorkflowsBytes, err := RetrieveWorkflows(c)
	if err != nil {
		return err
	}

	publishedWorkflows := []WorkflowStub{}
	err = json.Unmarshal(publishedWorkflowsBytes, &publishedWorkflows)
	if err != nil {
		return fmt.Errorf("error unmarshalling published workflows: %s", err)
	}

	var uploadedWorkflow *WorkflowStub
	for i := range publishedWorkflows {
		if publishedWorkflows[i].Name == workflowStub.Name {
			uploadedWorkflow = &publishedWorkflows[i]
		}
	}

	if uploadedWorkflow == nil {
		return errors.New("workflow was not successfully uploaded to server")
	}

	return nil
}

func RetrieveWorkflows(c config.Cpi) ([]byte, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/library", c.ApiServer)
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

func WorkflowFetcher(c config.Cpi, nodeID string, workflowID string) (WorkflowResponse, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows", c.ApiServer, nodeID)
	resp, err := http.Get(url)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("Error requesting active workflows on node at url: %s, msg: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return WorkflowResponse{}, fmt.Errorf("Failed retrieving active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}

	body, err := ioutil.ReadAll(resp.Body)
	var workflows []WorkflowResponse
	err = json.Unmarshal(body, &workflows)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("Error unmarshalling active workflows: %s", err)
	}

	var w *WorkflowResponse
	for i := range workflows {
		if workflows[i].ID == workflowID {
			w = &workflows[i]
			break
		}
	}

	if w == nil {
		return WorkflowResponse{}, fmt.Errorf("could not find workflow with name: %s on node: %s", workflowID, nodeID)
	}

	return *w, nil
}

func WorkflowPoster(c config.Cpi, nodeID string, req RunWorkflowRequestBody) (WorkflowResponse, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/", c.ApiServer, nodeID)
	body, err := json.Marshal(req)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error marshalling workflow request body, %s", err)
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error building http request to run workflow, %s", err)
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error running workflow at url %s", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return WorkflowResponse{}, fmt.Errorf("Failed running workflow at url: %s with status: %s, message: %s, body: %s", url, resp.Status, string(msg), string(body))
	}

	wfRespBytes, err := ioutil.ReadAll(resp.Body)
	log.Debug(fmt.Sprintf("body response is %s", string(wfRespBytes)))
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error reading workflow response body %s", err)
	}

	workflowResp := WorkflowResponse{}
	err = json.Unmarshal(wfRespBytes, &workflowResp)
	if err != nil {
		return WorkflowResponse{}, fmt.Errorf("error unmarshalling /common/node/workflows response %s", err)
	}

	log.Debug(fmt.Sprintf("workflow response %v", workflowResp))

	return workflowResp, nil
}

func RunWorkflow(poster workflowPosterFunc, fetcher workflowFetcherFunc, c config.Cpi, nodeID string, req RunWorkflowRequestBody) error {
	workflowResponse, err := poster(c, nodeID, req)
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
			wr, err := fetcher(c, nodeID, workflowResponse.ID)
			if err != nil {
				return fmt.Errorf("Unable to fetch workflow status: %s", err)
			}

			for key, value := range wr.Tasks {
				log.Debug(fmt.Sprintf("task is %s => %v", key, value))
			}

			log.Debug(fmt.Sprintf("workflow: %s with status: %s and pending tasks: %d", wr.Name, wr.Status, len(wr.PendingTasks)))

			switch wr.Status {
			case workflowValidStatus:
				if len(wr.PendingTasks) == 0 {
					log.Info(fmt.Sprintf("workflow: %s completed with valid state against node: %s", req.Name, nodeID))
					return nil
				}
				log.Debug(fmt.Sprintf("workflow: %s is still running against node: %s", req.Name, nodeID))
				continue
			case worfklowSuccessfulStatus:
				log.Info(fmt.Sprintf("workflow: %s completed successfully against node: %s", req.Name, nodeID))
				return nil
			case workflowFailedStatus:
				return fmt.Errorf("workflow: %s failed against node: %s", req.Name, nodeID)
			case workflowCancelledStatus:
				log.Info(fmt.Sprintf("workflow: %s was cancelled against node: %s", req.Name, nodeID))
				return nil
			default:
				return fmt.Errorf("workflow: %s has unexpected status %s on node: %s", req.Name, wr.Status, nodeID)
			}
		}
	}
}

func KillActiveWorkflow(c config.Cpi, nodeID string) error {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/active", c.ApiServer, nodeID)
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

func getActiveWorkflows(c config.Cpi, nodeID string) ([]WorkflowResponse, error) {
	var workflows []WorkflowResponse

	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/active", c.ApiServer, nodeID)
	resp, err := http.Get(url)
	if err != nil {
		return []WorkflowResponse{}, fmt.Errorf("Error requesting active workflows on node at url: %s, msg: %s", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return []WorkflowResponse{}, fmt.Errorf("Failed retrieving active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &workflows)
	if err != nil {
		return []WorkflowResponse{}, fmt.Errorf("Error unmarshalling active workflows: %s", err)
	}

	return workflows, nil
}
