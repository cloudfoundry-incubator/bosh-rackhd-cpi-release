package workflows

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
)

const (
	requiredTaskLength = 3
)

func BootstrappingTasksExist(c config.Cpi) error {
	tasks, err := onrackhttp.RetrieveTasks(c)
	if err != nil {
		log.Printf("unable to retrieve tasks: %s", err)
		return fmt.Errorf("unable to retrieve tasks: %s", err)
	}

	foundTasks := map[string]onrackhttp.Task{}

	for i := range tasks {
		switch tasks[i].Name {
		case SetPxeBootTaskName:
			foundTasks[SetPxeBootTaskName] = tasks[i]
		case RebootNodeTaskName:
			foundTasks[RebootNodeTaskName] = tasks[i]
		case BootstrapUbuntuTaskName:
			foundTasks[BootstrapUbuntuTaskName] = tasks[i]
		}

		if len(foundTasks) == requiredTaskLength {
			break
		}
	}

	if len(foundTasks) != requiredTaskLength {
		log.Printf("Did not find the expected number of required bootstrapping tasks: %d, found %v", requiredTaskLength, foundTasks)
		return fmt.Errorf("Did not find the expected number of required bootstrapping tasks: %d, found %v", requiredTaskLength, foundTasks)
	}

	if !setPxeBootTaskIsExpected(foundTasks[SetPxeBootTaskName]) {
		log.Printf("Set PXE boot task has unexpected form: %v\n", foundTasks[SetPxeBootTaskName])
		return fmt.Errorf("Set PXE boot task has unexpected form: %v", foundTasks[SetPxeBootTaskName])
	}

	if !rebootNodeTaskIsExpected(foundTasks[RebootNodeTaskName]) {
		log.Printf("Reboot node task has unexpected form: %v\n", foundTasks[RebootNodeTaskName])
		return fmt.Errorf("Reboot node task has unexpected form: %v", foundTasks[RebootNodeTaskName])
	}

	if !bootstrapUbuntuIsExpected(foundTasks[BootstrapUbuntuTaskName]) {
		log.Printf("Bootstrap Ubuntu task has unexpected form: %v\n", foundTasks[BootstrapUbuntuTaskName])
		return fmt.Errorf("Bootstrap Ubuntu task has unexpected form: %v", foundTasks[BootstrapUbuntuTaskName])
	}

	return nil
}

func setPxeBootTaskIsExpected(t onrackhttp.Task) bool {
	expectedTask := onrackhttp.Task{}
	err := json.Unmarshal(setPxeBootTemplate, &expectedTask)
	if err != nil {
		log.Printf("Error unmarshalling setPxeBootTemplate: %s", err)
		return false
	}

	if !reflect.DeepEqual(t, expectedTask) {
		return false
	}

	return true
}

func rebootNodeTaskIsExpected(t onrackhttp.Task) bool {
	expectedTask := onrackhttp.Task{}
	err := json.Unmarshal(rebootNodeTemplate, &expectedTask)
	if err != nil {
		log.Printf("Error unmarshalling rebootNodeTemplate: %s", err)
		return false
	}

	if !reflect.DeepEqual(t, expectedTask) {
		return false
	}

	return true
}

func bootstrapUbuntuIsExpected(t onrackhttp.Task) bool {
	expectedTask := onrackhttp.Task{}
	err := json.Unmarshal(bootstrapUbuntuTaskTemplate, &expectedTask)
	if err != nil {
		log.Printf("Error unmarshalling bootstrapUbuntuTaskTemplate: %s", err)
		return false
	}

	if !reflect.DeepEqual(t, expectedTask) {
		return false
	}

	return true
}
