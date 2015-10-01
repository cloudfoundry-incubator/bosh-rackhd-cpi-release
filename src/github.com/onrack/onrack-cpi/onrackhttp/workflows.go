package onrackhttp

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

func PublishWorkflow(c config.Cpi, workflowBytes []byte) error {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows", c.ApiServer)

	request, err := http.NewRequest("PUT", url, bytes.NewReader(workflowBytes))
	if err != nil {
		log.Error(fmt.Sprintf("error building http request: %s\n", err))
		return fmt.Errorf("error building http request: %s", err)
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(fmt.Sprintf("error sending publishing workflow to %s\n", url))
		return fmt.Errorf("error sending publishing workflow to %s", url)
	}
	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error reading response body: %s\n", err))
		return fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Error(fmt.Sprintf("error response code is %d: %s\n", resp.StatusCode, string(msg)))
		return fmt.Errorf("Failed publishing workflows with status: %s", resp.Status)
	}

	workflowStub := WorkflowStub{}
	err = json.Unmarshal(workflowBytes, &workflowStub)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling workflow: %s\n", err))
		return fmt.Errorf("error unmarshalling workflow: %s\n", err)
	}

	publishedWorkflowsBytes, err := RetrieveWorkflows(c)
	if err != nil {
		return err
	}

	publishedWorkflows := []WorkflowStub{}
	err = json.Unmarshal(publishedWorkflowsBytes, &publishedWorkflows)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling published workflows: %s\n", err))
		return fmt.Errorf("error unmarshalling published workflows: %s\n", err)
	}

	var uploadedWorkflow *WorkflowStub
	for i := range publishedWorkflows {
		if publishedWorkflows[i].Name == workflowStub.Name {
			uploadedWorkflow = &publishedWorkflows[i]
		}
	}

	if uploadedWorkflow == nil {
		log.Error("workflow was not successfully uploaded to server")
		return errors.New("workflow was not successfully uploaded to server")
	}

	return nil
}

func RetrieveWorkflows(c config.Cpi) ([]byte, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/library", c.ApiServer)
	resp, err := http.Get(url)
	if err != nil {
		log.Error(fmt.Sprintf("Error: %s\n", err))
		return nil, fmt.Errorf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error reading response body: %s\n", err))
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Error(fmt.Sprintf("error retrieving tasks: response code is %d: %s\n", resp.StatusCode, string(body)))
		return nil, fmt.Errorf("Failed retrieving workflows with status: %s, message: %s", resp.Status, string(body))
	}

	return body, nil
}

func FetchWorkflowImpl(c config.Cpi, nodeID string, workflowID string) (WorkflowResponse, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows", c.ApiServer, nodeID)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error requesting active workflows on node at url: %s, msg: %s", url, err)
		return WorkflowResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Failed retrieving active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
		return WorkflowResponse{}, fmt.Errorf("Failed retrieving active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}

	body, err := ioutil.ReadAll(resp.Body)
	var workflows []WorkflowResponse
	err = json.Unmarshal(body, &workflows)
	if err != nil {
		log.Printf("Error unmarshalling active workflows: %s", err)
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
		log.Printf("could not find workflow with name: %s on node: %s", workflowID, nodeID)
		return WorkflowResponse{}, fmt.Errorf("could not find workflow with name: %s on node: %s", workflowID, nodeID)
	}

	return *w, nil
}

type FetchWorkflow func(config.Cpi, string, string) (WorkflowResponse, error)

func RunWorkflow(c config.Cpi, nodeID string, req RunWorkflowRequestBody, fetch FetchWorkflow) (err error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/", c.ApiServer, nodeID)
	body, err := json.Marshal(req)
	if err != nil {
		log.Error("error marshalling workflow request body")
		return
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Error("error building http request to run workflow")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(fmt.Sprintf("error running workflow at url %s", url))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Error(fmt.Sprintf("error response code is %d: %s, with body: %s", resp.StatusCode, string(msg), string(body)))
		return fmt.Errorf("Failed running workflow at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}

	wfRespBytes, err := ioutil.ReadAll(resp.Body)
	log.Debug(fmt.Sprintf("body response is %s", string(wfRespBytes)))
	if err != nil {
		log.Error(fmt.Sprintf("error reading workflow response body %s", err))
		return fmt.Errorf("error reading workflow response body %s", err)
	}

	workflowResp := WorkflowResponse{}
	err = json.Unmarshal(wfRespBytes, &workflowResp)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling /common/node/workflows response %s", err))
		return fmt.Errorf("error unmarshalling /common/node/workflows response %s", err)
	}

	log.Info(fmt.Sprintf("workflow response %v", workflowResp))
	timeoutChan := time.NewTimer(time.Second * c.RunWorkflowTimeoutSeconds).C
	retryChan := time.NewTicker(time.Second * 5).C

	for {
		select {
		case <-timeoutChan:
			err := killActiveWorkflows(c, nodeID)
			if err != nil {
				log.Error(fmt.Sprintf("Could not abort timed out workflow on node: %s\n", nodeID))
				return fmt.Errorf("Could not abort timed out workflow on node: %s", nodeID)
			}

			log.Info(fmt.Sprintf("Timed out running workflow: %s on node: %s", req.Name, nodeID))
			return fmt.Errorf("Timed out running workflow: %s on node: %s", req.Name, nodeID)
		case <-retryChan:
			workflowResponse, err := fetch(c, nodeID, workflowResp.ID)
			if err != nil {
				log.Error(fmt.Sprintf("Unable to fetch workflow status: %s\n", err))
				return err
			}

			switch workflowResponse.Status {
			case workflowValidStatus:
				if len(workflowResponse.PendingTasks) == 0 {
					log.Info(fmt.Sprintf("workflow: %s completed with valid state against node: %s", req.Name, nodeID))
					return nil
				}
				log.Debug(fmt.Sprintf("workflow: %s is still running against node: %s", req.Name, nodeID))
				continue
			case worfklowSuccessfulStatus:
				log.Info(fmt.Sprintf("workflow: %s completed successfully against node: %s", req.Name, nodeID))
				return nil
			case workflowFailedStatus:
				log.Error(fmt.Sprintf("workflow: %s failed against node: %s", req.Name, nodeID))
				return fmt.Errorf("workflow: %s failed against node: %s", req.Name, nodeID)
			case workflowCancelledStatus:
				log.Error(fmt.Sprintf("workflow: %s was cancelled against node: %s", req.Name, nodeID))
				return nil
			default:
				log.Error(fmt.Sprintf("workflow: %s has unexpected status on node: %s", req.Name, nodeID))
				return fmt.Errorf("workflow: %s has unexpected status %s on node: %s", req.Name, workflowResponse.Status, nodeID)
			}
		}
	}
}

func killActiveWorkflows(c config.Cpi, nodeID string) error {
	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/active", c.ApiServer, nodeID)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Error(fmt.Sprintf("error: %s  building http request to delete active workflows against node: %s\n", err, nodeID))
		return fmt.Errorf("error: %s building http request to delete active workflows against node: %s", err, nodeID)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(fmt.Sprintf("Error: %s deleting active workflows on node: %s\n", err, nodeID))
		return fmt.Errorf("Error: %s deleting active workflows on node: %s", err, nodeID)
	}

	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Error: %s reading response body from abort workflows request against node: %s\n", err, nodeID))
		return fmt.Errorf("Error: %s reading response body from abort workflows request against node: %s", err, nodeID)
	}

	if resp.StatusCode != 200 {
		log.Error(fmt.Sprintf("Failed deleting active workflows against node: %s with status: %s, message: %s", nodeID, resp.Status, string(msg)))
		return fmt.Errorf("Failed deleting active workflows against node: %s with status: %s, message: %s", nodeID, resp.Status, string(msg))
	}

	return nil
}

func getActiveWorkflows(c config.Cpi, nodeID string) ([]WorkflowResponse, error) {
	var workflows []WorkflowResponse

	url := fmt.Sprintf("http://%s:8080/api/1.1/nodes/%s/workflows/active", c.ApiServer, nodeID)
	resp, err := http.Get(url)
	if err != nil {
		log.Error(fmt.Sprintf("Error requesting active workflows on node at url: %s, msg: %s", url, err))
		return []WorkflowResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Error(fmt.Sprintf("Failed retrieving active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg)))
		return []WorkflowResponse{}, fmt.Errorf("Failed retrieving active workflows at url: %s with status: %s, message: %s", url, resp.Status, string(msg))
	}

	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &workflows)
	if err != nil {
		log.Error(fmt.Sprintf("Error unmarshalling active workflows: %s", err))
		return []WorkflowResponse{}, fmt.Errorf("Error unmarshalling active workflows: %s", err)
	}

	return workflows, nil
}

const (
	OnrackReserveVMGraphName = "Graph.CF.ReserveVM"
	OnrackCreateVMGraphName  = "Graph.BOSH.ProvisionNode"
	OnrackDeleteVMGraphName  = "Graph.CF.DeleteVM"
	OnrackEnvPath            = "/var/vcap/bosh/onrack-cpi-agent-env.json"
	OnrackRegistryPath       = "/var/vcap/bosh/agent.json"
	DefaultUnusedName        = "UPLOADED_BY_ONRACK_CPI"
)

const (
	workflowValidStatus      = "valid"
	worfklowSuccessfulStatus = "succeeded"
	workflowFailedStatus     = "failed"
	workflowCancelledStatus  = "cancelled"
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
