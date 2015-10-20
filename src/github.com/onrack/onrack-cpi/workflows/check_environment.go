package workflows

import (
	"encoding/json"
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"

	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
)

const (
	requiredTaskLength = 3
)

type bootstrapTask struct {
	*onrackapi.TaskStub
	Options    json.RawMessage `json:"options"`
	Properties json.RawMessage `json:"properties"`
}

func BootstrappingTasksExist(c config.Cpi) error {
	tasksBytes, err := onrackapi.RetrieveTasks(c)
	if err != nil {
		return fmt.Errorf("unable to retrieve tasks: %s", err)
	}

	tasks := []bootstrapTask{}
	err = json.Unmarshal(tasksBytes, &tasks)

	foundTasks := map[string]interface{}{}

	for i := range tasks {
		switch tasks[i].Name {
		case SetPxeBootTaskName:
			options := obmServiceOptions{}
			properties := setPxeBootTaskProperties{}

			err = json.Unmarshal(tasks[i].Options, &options)
			if err != nil {
				return fmt.Errorf("error unmarshalling pxe boot task options: %s", err)
			}

			err = json.Unmarshal(tasks[i].Properties, &properties)
			if err != nil {
				return fmt.Errorf("error unmarshalling pxe boot task properties: %s", err)
			}

			t := setPxeBootTask{TaskStub: tasks[i].TaskStub,
				setPxeBootTaskOptionsContainer:    &setPxeBootTaskOptionsContainer{},
				setPxeBootTaskPropertiesContainer: &setPxeBootTaskPropertiesContainer{},
			}

			t.Options = options
			t.Properties = properties
			foundTasks[SetPxeBootTaskName] = t
		case RebootNodeTaskName:
			options := obmServiceOptions{}
			properties := rebootNodeTaskProperties{}

			err = json.Unmarshal(tasks[i].Options, &options)
			if err != nil {
				return fmt.Errorf("error unmarshalling reboot task options: %s", err)
			}

			err = json.Unmarshal(tasks[i].Properties, &properties)
			if err != nil {
				return fmt.Errorf("error unmarshalling pxe boot task properties: %s", err)
			}

			t := rebootNodeTask{
				TaskStub:                          tasks[i].TaskStub,
				rebootNodeTaskOptionsContainer:    &rebootNodeTaskOptionsContainer{},
				rebootNodeTaskPropertiesContainer: &rebootNodeTaskPropertiesContainer{},
			}

			t.Options = options
			t.Properties = properties
			foundTasks[RebootNodeTaskName] = t
		case BootstrapUbuntuTaskName:
			options := bootstrapUbuntuTaskOptions{}
			properties := bootstrapUbuntuTaskProperties{}

			err = json.Unmarshal(tasks[i].Options, &options)
			if err != nil {
				return fmt.Errorf("error unmarshalling bootstrap ubuntu task options: %s", err)
			}

			err = json.Unmarshal(tasks[i].Properties, &properties)
			if err != nil {
				return fmt.Errorf("error unmarshalling bootstrap ubuntu task properties: %s", err)
			}

			t := bootstrapUbuntuTask{
				TaskStub: tasks[i].TaskStub,
				bootstrapUbuntuTaskOptionsContainer:    &bootstrapUbuntuTaskOptionsContainer{},
				bootstrapUbuntuTaskPropertiesContainer: &bootstrapUbuntuTaskPropertiesContainer{},
			}

			t.Options = options
			t.Properties = properties
			foundTasks[BootstrapUbuntuTaskName] = t
		}

		if len(foundTasks) == requiredTaskLength {
			break
		}
	}

	if len(foundTasks) != requiredTaskLength {
		return fmt.Errorf("Did not find the expected number of required bootstrapping tasks: %d, found %v", requiredTaskLength, foundTasks)
	}

	if !setPxeBootTaskIsExpected(foundTasks[SetPxeBootTaskName].(setPxeBootTask)) {
		return fmt.Errorf("Set PXE boot task has unexpected form: %v", foundTasks[SetPxeBootTaskName])
	}

	if !rebootNodeTaskIsExpected(foundTasks[RebootNodeTaskName].(rebootNodeTask)) {
		return fmt.Errorf("Reboot node task has unexpected form: %v", foundTasks[RebootNodeTaskName])
	}

	if !bootstrapUbuntuIsExpected(foundTasks[BootstrapUbuntuTaskName].(bootstrapUbuntuTask)) {
		return fmt.Errorf("Bootstrap Ubuntu task has unexpected form: %v", foundTasks[BootstrapUbuntuTaskName])
	}

	return nil
}

func setPxeBootTaskIsExpected(t setPxeBootTask) bool {
	expectedTask := setPxeBootTask{}

	err := json.Unmarshal(setPxeBootTemplate, &expectedTask)
	if err != nil {
		log.Error(fmt.Sprintf("Error unmarshalling setPxeBootTemplate: %s", err))
		return false
	}

	if !reflect.DeepEqual(t, expectedTask) {
		log.Error(fmt.Sprintf("actual %v, expected %v", t, expectedTask))
		return false
	}

	return true
}

func rebootNodeTaskIsExpected(t rebootNodeTask) bool {
	expectedTask := rebootNodeTask{}
	err := json.Unmarshal(rebootNodeTemplate, &expectedTask)
	if err != nil {
		log.Error(fmt.Sprintf("Error unmarshalling rebootNodeTemplate: %s", err))
		return false
	}

	if !reflect.DeepEqual(t, expectedTask) {
		log.Error(fmt.Sprintf("actual %v, expected %v", t, expectedTask))
		return false
	}

	return true
}

func bootstrapUbuntuIsExpected(t bootstrapUbuntuTask) bool {
	expectedTask := bootstrapUbuntuTask{}
	err := json.Unmarshal(bootstrapUbuntuTaskTemplate, &expectedTask)
	if err != nil {
		log.Error(fmt.Sprintf("Error unmarshalling bootstrapUbuntuTaskTemplate: %s", err))
		return false
	}

	if !reflect.DeepEqual(t, expectedTask) {
		log.Error(fmt.Sprintf("actual %v, expected %v", t, expectedTask))
		return false
	}

	return true
}
