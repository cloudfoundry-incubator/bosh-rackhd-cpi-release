package rackhdapi

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "time"

  log "github.com/Sirupsen/logrus"

  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/helpers"
  "github.com/rackhd/rackhd-cpi/models"
)

type workflowFetcherFunc func(config.Cpi, string) (models.WorkflowResponse, error)

type workflowPosterFunc func(config.Cpi, string, models.RunWorkflowRequestBody) (models.WorkflowResponse, error)

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

  graph := models.Graph{}
  err = json.Unmarshal(graphBytes, &graph)
  if err != nil {
    return fmt.Errorf("error unmarshalling graph: %s", err)
  }
  log.Debug("workflow received after publishing: %s", string(graphBytes))

  _, err = RetrieveGraph(c, graph.Name)
  return err
}

func RetrieveGraph(c config.Cpi, graphName string) (models.Graph, error) {
  url := fmt.Sprintf("%s/api/2.0/workflows/graphs/%s", c.ApiServer, graphName)
  respBody, err := helpers.MakeRequest(url, "GET", 200, nil)
  if err != nil {
    return models.Graph{}, err
  }

  graphs := []models.Graph{}
  err = json.Unmarshal(respBody, &graphs)
  if err != nil {
    return models.Graph{}, err
  }

  if len(graphs) != 1 {
    return models.Graph{}, fmt.Errorf("incorrect number of graphs returned: %d", len(graphs))
  }
  return graphs[0], nil
}

func WorkflowFetcher(c config.Cpi, workflowIntanceID string) (models.WorkflowResponse, error) {
  url := fmt.Sprintf("%s/api/2.0/workflows/%s", c.ApiServer, workflowIntanceID)
  respBody, err := helpers.MakeRequest(url, "GET", 200, nil)
  if err != nil {
    return models.WorkflowResponse{}, err
  }

  var workflow models.WorkflowResponse
  err = json.Unmarshal(respBody, &workflow)
  if err != nil {
    return models.WorkflowResponse{}, err
  }

  return workflow, nil
}

// WorkflowPoster runs a graph as workflow
func WorkflowPoster(c config.Cpi, nodeID string, req models.RunWorkflowRequestBody) (models.WorkflowResponse, error) {
  reqBody, err := json.Marshal(req)
  if err != nil {
    return models.WorkflowResponse{}, fmt.Errorf("error marshalling workflow request body, %s", err)
  }
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/workflows", c.ApiServer, nodeID)
  respBody, err := helpers.MakeRequest(url, "POST", 201, reqBody)
  if err != nil {
    return models.WorkflowResponse{}, err
  }

  workflowResp := models.WorkflowResponse{}
  err = json.Unmarshal(respBody, &workflowResp)
  if err != nil {
    return models.WorkflowResponse{}, fmt.Errorf("error unmarshalling returned workflow response %s", err)
  }

  log.Info("Workflow post successful")
  return workflowResp, nil
}

func RunWorkflow(poster workflowPosterFunc, fetcher workflowFetcherFunc, c config.Cpi, nodeID string, req models.RunWorkflowRequestBody) error {
  postedWorkflow, err := poster(c, nodeID, req)
  if err != nil {
    fmt.Printf("formatadsfasd: %+v", err)

    return fmt.Errorf("Failed to post workflow: %s", err)
  }

  timeoutChan := time.NewTimer(time.Second * c.RunWorkflowTimeoutSeconds).C
  retryChan := time.NewTicker(time.Second * 3).C

  for {
    select {
    case <-timeoutChan:
      err := KillActiveWorkflow(c, nodeID)
      if err != nil {
        return fmt.Errorf("Could not kill timed out workflow on node: %s, error: %s", nodeID, err)
      }
      return fmt.Errorf("Timed out running workflow: %s on node: %s", req.Name, nodeID)

    case <-retryChan:
      wr, err := fetcher(c, postedWorkflow.InstanceID)
      if err != nil {
        return fmt.Errorf("Unable to fetch workflow status: %s", err)
      }

      switch wr.Status {
      case models.WorkflowRunningStatus:
        log.Info(fmt.Sprintf("workflow: %s is running against node: %s", wr.InstanceID, nodeID))
        continue
      case models.WorkflowSuccessfulStatus:
        log.Info(fmt.Sprintf("workflow: %s completed successfully against node: %s", wr.InstanceID, nodeID))
        return nil
      case models.WorkflowFailedStatus:
        return fmt.Errorf("workflow: %s failed against node: %s", wr.InstanceID, nodeID)
      case models.WorkflowCancelledStatus:
        log.Info(fmt.Sprintf("workflow: %s was cancelled against node: %s", wr.InstanceID, nodeID))
        return nil
      case models.WorkflowPendingStatus:
        log.Info(fmt.Sprintf("workflow: %s is pending on node: %s", wr.InstanceID, nodeID))
        continue
      default:
        return fmt.Errorf("workflow: %s has unexpected status '%s' on node: %s", wr.InstanceID, wr.Status, nodeID)
      }
    }
  }
}

func KillActiveWorkflow(c config.Cpi, nodeID string) error {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/workflows/action", c.ApiServer, nodeID)
  _, err := helpers.MakeRequest(url, "PUT", 202, []byte("{\"command\": \"cancel\",\"options\": {}}"))
  return err
}

func GetActiveWorkflow(c config.Cpi, nodeID string) (models.WorkflowResponse, error) {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/workflows", c.ApiServer, nodeID)
  req, err := http.NewRequest("GET", url, nil)
  params := req.URL.Query()
  params.Add("active", "true")
  req.URL.RawQuery = params.Encode()

  respBody, err := helpers.MakeConfigedRequest(req, []int{200})
  if err != nil {
    return models.WorkflowResponse{}, err
  }

  workflows := []models.WorkflowResponse{}
  err = json.Unmarshal(respBody, &workflows)
  if err != nil {
    return models.WorkflowResponse{}, err
  }

  if len(workflows) != 1 {
    return models.WorkflowResponse{}, fmt.Errorf("incorrect number of active workflows returned: %d", len(workflows))
  }
  return workflows[0], nil
}
