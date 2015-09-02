package action

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
	"net/http"
	"errors"
	"strings"
)

type DeleteVM struct {
	APIServer string
	logger boshlog.Logger
	logTag string
}

func NewDeleteVM(APIServer string, logger boshlog.Logger) DeleteVM {
	return DeleteVM{
		APIServer: APIServer,
		logger: logger,
		logTag: "delete-vm",
	}
}

func (a DeleteVM) Run(vmCID VMCID) (interface{}, error) {
	deleteUrl := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows?name=Graph.CF.DeleteVM", a.APIServer, vmCID)
	a.logger.Info(a.logTag, "Delete Url is '%s'", deleteUrl)

	resp, err := http.Post(deleteUrl, "", strings.NewReader(""))


	if err != nil {
		return nil, errors.New("Error deleting vm")
	}

	a.logger.Info(a.logTag, "Response status is '%s'", resp.Status)
	if (resp.StatusCode != http.StatusCreated) {
		return nil, errors.New("Error creating workflow")
	}

	//succeeded in creating workflow, so wait for the workflow to finish before returning
	for a.isWorkFlowActive(string(vmCID)) {
		time.Sleep(30 * time.Second)
	}

	return nil, nil
}

func (a DeleteVM) isWorkFlowActive(machineID string) bool {
	a.logger.Info(a.logTag, "Checking workflow actively")
	workFlowUrl := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows/active", a.APIServer, machineID)
	resp, err := http.Get(workFlowUrl)

	if (err != nil) {
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var workflow Workflow
	err = json.Unmarshal(body, &workflow)
	if err != nil {
		return false
	}

	if workflow.Status != nil {
		return true
	}

	return false
}
