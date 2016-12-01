package rackhdapi

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
)

func PublishTask(c config.Cpi, taskBytes []byte) error {
	url := fmt.Sprintf("%s/api/2.0/workflows/tasks", c.ApiServer)
	respBody, err := helpers.MakeRequest(url, "PUT", 201, taskBytes)
	if err != nil {
		return err
	}

	task := models.Task{}
	err = json.Unmarshal(respBody, &task)
	if err != nil {
		return fmt.Errorf("error unmarshalling task: %s", err)
	}
	log.Debug("task to publish: %+v", task)

	publishedTask, err := RetrieveTask(c, task.Name)
	if err != nil {
		return err
	}
	log.Debug("published task: %+v", publishedTask)

	if publishedTask.Name == task.Name {
		return nil
	}
	return fmt.Errorf("Task was not successfully uploaded to server!\n")
}

func RetrieveTask(c config.Cpi, taskName string) (models.Task, error) {
	url := fmt.Sprintf("%s/api/2.0/workflows/tasks/%s", c.ApiServer, taskName)
	respBody, err := helpers.MakeRequest(url, "GET", 200, nil)
	if err != nil {
		return models.Task{}, err
	}

	tasks := []models.Task{}
	err = json.Unmarshal(respBody, &tasks)
	if err != nil {
		return models.Task{}, err
	}

	if len(tasks) == 0 {
		return models.Task{}, fmt.Errorf("could not find %s", taskName)
	}
	return tasks[0], nil
}

func DeleteTask(c config.Cpi, taskName string) error{
	log.Info(fmt.Sprintf("deleting task %s", taskName))
	url := fmt.Sprintf("%s/api/2.0/workflows/tasks/%s", c.ApiServer, taskName)
	_, err := helpers.MakeRequest(url, "DELETE", 204, nil)
	if err != nil {
		return fmt.Errorf("error deleting task %s", err)
	}

	return nil
}