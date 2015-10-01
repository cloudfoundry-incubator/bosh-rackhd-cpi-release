package onrackhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/onrack/onrack-cpi/config"
)

func PublishTask(c config.Cpi, taskBytes []byte) error {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/tasks", c.ApiServer)
	request, err := http.NewRequest("PUT", url, bytes.NewReader(taskBytes))
	if err != nil {
		log.Error("error building publish task request")
		return errors.New("error building publish task request")
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Error(fmt.Sprintf("error sending PUT request to %s", c.ApiServer))
		return fmt.Errorf("error sending PUT request to %s", c.ApiServer)
	}

	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("error reading response body: %s\n", err))
		return fmt.Errorf("error reading response body: %s\n", err)
	}

	if resp.StatusCode != 200 {
		log.Error(fmt.Sprintf("error publishing task: response code is %d: %s", resp.StatusCode, string(msg)))
		return fmt.Errorf("Failed publishing task with status: %s, message: %s", resp.Status, string(msg))
	}

	taskStub := TaskStub{}
	err = json.Unmarshal(taskBytes, &taskStub)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling task: %s\n", err))
		return fmt.Errorf("error unmarshalling task: %s\n", err)
	}

	publishedTaskBytes, err := RetrieveTasks(c)
	if err != nil {
		return err
	}

	publishedTaskStubs := []TaskStub{}
	err = json.Unmarshal(publishedTaskBytes, &publishedTaskStubs)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling published tasks: %s\n", err))
		return fmt.Errorf("error unmarshalling published tasks: %s\n", err)
	}

	var uploadedTaskStub *TaskStub
	for i := range publishedTaskStubs {
		if publishedTaskStubs[i].Name == taskStub.Name {
			uploadedTaskStub = &publishedTaskStubs[i]
		}
	}

	if uploadedTaskStub == nil {
		log.Error("task was not successfully uploaded to server")
		return errors.New("task was not successfully uploaded to server")
	}

	return nil
}

func RetrieveTasks(c config.Cpi) ([]byte, error) {
	url := fmt.Sprintf("http://%s:8080/api/1.1/workflows/tasks/library", c.ApiServer)
	resp, err := http.Get(url)
	if err != nil {
		log.Error(fmt.Sprintf("Error: %s", err))
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
		return nil, fmt.Errorf("error retrieving tasks: response code is %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

type TaskStub struct {
	Name           string `json:"injectableName"`
	UnusedName     string `json:"friendlyName"`
	ImplementsTask string `json:"implementsTask,omitempty"`
}

type WorkflowTask struct {
	TaskName      string            `json:"taskName"`
	Label         string            `json:"label"`
	WaitOn        map[string]string `json:"waitOn,omitempty"`
	IgnoreFailure bool              `json:"ignoreFailure,omitempty"`
}

type TaskResponse struct {
	Name  string `json:"name"`
	State string `json:"state"`
}
