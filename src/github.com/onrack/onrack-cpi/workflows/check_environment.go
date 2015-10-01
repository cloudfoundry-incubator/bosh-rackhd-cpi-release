package workflows

import (
	"encoding/json"
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"

	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
)

const (
	requiredTaskLength = 3
)

type bootstrapTask struct {
	*onrackhttp.TaskStub
	Options    json.RawMessage `json:"options"`
	Properties json.RawMessage `json:"properties"`
}

func BootstrappingTasksExist(c config.Cpi) error {
	tasksBytes, err := onrackhttp.RetrieveTasks(c)
	if err != nil {
		log.Error(fmt.Sprintf("unable to retrieve tasks: %s", err))
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
				log.Error(fmt.Sprintf("error unmarshalling pxe boot task options: %s\n", err))
				return fmt.Errorf("error unmarshalling pxe boot task options: %s\n", err)
			}

			err = json.Unmarshal(tasks[i].Properties, &properties)
			if err != nil {
				log.Error(fmt.Sprintf("error unmarshalling pxe boot task properties: %s\n", err))
				return fmt.Errorf("error unmarshalling pxe boot task properties: %s\n", err)
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
				log.Error(fmt.Sprintf("error unmarshalling reboot task options: %s\n", err))
				return fmt.Errorf("error unmarshalling reboot task options: %s\n", err)
			}

			err = json.Unmarshal(tasks[i].Properties, &properties)
			if err != nil {
				log.Error(fmt.Sprintf("error unmarshalling pxe boot task properties: %s\n", err))
				return fmt.Errorf("error unmarshalling pxe boot task properties: %s\n", err)
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
				log.Error(fmt.Sprintf("error unmarshalling bootstrap ubuntu task options: %s\n", err))
				return fmt.Errorf("error unmarshalling bootstrap ubuntu task options: %s\n", err)
			}

			err = json.Unmarshal(tasks[i].Properties, &properties)
			if err != nil {
				log.Error(fmt.Sprintf("error unmarshalling bootstrap ubuntu task properties: %s\n", err))
				return fmt.Errorf("error unmarshalling bootstrap ubuntu task properties: %s\n", err)
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
		log.Error(fmt.Sprintf("Did not find the expected number of required bootstrapping tasks: %d, found %v", requiredTaskLength, foundTasks))
		return fmt.Errorf("Did not find the expected number of required bootstrapping tasks: %d, found %v", requiredTaskLength, foundTasks)
	}

	if !setPxeBootTaskIsExpected(foundTasks[SetPxeBootTaskName].(setPxeBootTask)) {
		log.Error(fmt.Sprintf("Set PXE boot task has unexpected form: %v\n", foundTasks[SetPxeBootTaskName]))
		return fmt.Errorf("Set PXE boot task has unexpected form: %v", foundTasks[SetPxeBootTaskName])
	}

	if !rebootNodeTaskIsExpected(foundTasks[RebootNodeTaskName].(rebootNodeTask)) {
		log.Error(fmt.Sprintf("Reboot node task has unexpected form: %v\n", foundTasks[RebootNodeTaskName]))
		return fmt.Errorf("Reboot node task has unexpected form: %v", foundTasks[RebootNodeTaskName])
	}

	if !bootstrapUbuntuIsExpected(foundTasks[BootstrapUbuntuTaskName].(bootstrapUbuntuTask)) {
		log.Error(fmt.Sprintf("Bootstrap Ubuntu task has unexpected form: %v\n", foundTasks[BootstrapUbuntuTaskName]))
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
		return false
	}

	return true
}
