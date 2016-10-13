package rackhdapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/models"
)

func PublishTask(c config.Cpi, taskBytes []byte) error {
	url := fmt.Sprintf("%s/api/2.0/workflows/tasks", c.ApiServer)
	request, err := http.NewRequest("PUT", url, bytes.NewReader(taskBytes))
	if err != nil {
		return errors.New("error building publish task request")
	}
	request.Header.Set("Content-Type", "application/json")
	fmt.Printf("\n\nrequest: %+v, body: %+v", request, string(taskBytes))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("error sending PUT request to %s", c.ApiServer)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("bodyddd: %+v", string(b))
		return fmt.Errorf("error publishing task; response status code: %s,\nresponse body: %+v", resp.Status, resp)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}

	task := models.Task{}
	err = json.Unmarshal(taskBytes, &task)
	if err != nil {
		return fmt.Errorf("error unmarshalling task: %s", err)
	}
	log.Debug("task to publish: %+v", task)

	publishedTaskBytes, err := RetrieveTasks(c)
	if err != nil {
		return err
	}

	publishedTasks := []models.Task{}
	err = json.Unmarshal(publishedTaskBytes, &publishedTasks)
	if err != nil {
		return fmt.Errorf("error unmarshalling published tasks: %s", err)
	}

	var uploadedTask *models.Task
	for i := range publishedTasks {
		if publishedTasks[i].Name == task.Name {
			uploadedTask = &publishedTasks[i]
		}
	}

	if uploadedTask == nil {
		return fmt.Errorf("Task was not successfully uploaded to server!\n Request: %+v\n Response: %+v", request, resp)
	}

	return nil
}

func RetrieveTasks(c config.Cpi) ([]byte, error) {
	url := fmt.Sprintf("%s/api/1.1/workflows/tasks/library", c.ApiServer)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error retrieving tasks: response code is %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
