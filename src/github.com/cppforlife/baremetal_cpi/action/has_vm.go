package action

import (
	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
	"fmt"
	"encoding/json"
	"io/ioutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	httpclient "github.com/cppforlife/baremetal_cpi/utils/httpclient"
	"errors"
)

type HasVM struct {
	vmFinder bwcvm.Finder
	APIServer string
	logger boshlog.Logger
	logTag string
}

func NewHasVM(vmFinder bwcvm.Finder, logger boshlog.Logger, APIServer string) HasVM {
	return HasVM{
		vmFinder: vmFinder,
		APIServer: APIServer,
		logger: logger,
		logTag: "has-vm",
	}
}

func (a HasVM) Run(vmCID VMCID) (bool, error) {
	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)
	resp, err := client.Get(fmt.Sprintf("http://%s:8080/api/common/nodes/%s", a.APIServer, vmCID))

	if (err != nil) {
		//maybe better/diff error handling
		return false, errors.New("Error Getting node")
	}
	defer client.Close()

	a.logger.Info(a.logTag, "The response status is '%s'", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.New("Error getting response body")
	}

	var node Node
	err = json.Unmarshal(body, &node)
	if err != nil {
		return false, errors.New("Unmarshalling Node Metadata")
	}

	if node.Reserved != nil && *(node.Reserved) == "true" {
		a.logger.Info(a.logTag, "The node's reserve status is '%s'", *(node.Reserved))
		return true, nil
	}

	return false, nil
}

type Node struct {
	Reserved *string `json:"reserved"`
}