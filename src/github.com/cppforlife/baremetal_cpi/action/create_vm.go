package action

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bwcstem "github.com/cppforlife/baremetal_cpi/stemcell"
	bwcvm "github.com/cppforlife/baremetal_cpi/vm"
	httpclient "github.com/cppforlife/baremetal_cpi/utils/httpclient"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"strings"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
	"net/http"
	"errors"
)

type CreateVM struct {
	stemcellFinder bwcstem.Finder
	vmCreator      bwcvm.Creator
	APIServer string
	agentOptions bwcvm.AgentOptions
	logger boshlog.Logger
	logTag string
}

type VMCloudProperties struct{}

type Environment map[string]interface{}

func NewCreateVM(stemcellFinder bwcstem.Finder, vmCreator bwcvm.Creator, APIServer string, agentOptions bwcvm.AgentOptions, logger boshlog.Logger) CreateVM {
	return CreateVM{
		stemcellFinder: stemcellFinder,
		vmCreator:      vmCreator,
		APIServer: APIServer,
		agentOptions: agentOptions,
		logger: logger,
		logTag: "create_vm",
	}
}

func (a CreateVM) Run(agentID string, stemcellCID StemcellCID, _ VMCloudProperties, networks Networks, _ []DiskCID, env Environment) (VMCID, error) {
	client := httpclient.NewHTTPClient(httpclient.DefaultClient, a.logger)
	machineID := "machine-id"
    //create agent and network env
	vmNetworks := networks.AsVMNetworks()
	vmEnv := bwcvm.Environment(env)
	agentEnv := bwcvm.NewAgentEnvForVM(agentID,machineID, vmNetworks, vmEnv, a.agentOptions)
	jsonBytes, err := json.Marshal(agentEnv)
	if err != nil {
		return "", bosherr.WrapError(err, "Marshalling agent env")
	}

	a.logger.Info(a.logTag, "Agent Json data '%s'", string(jsonBytes))

    //call api to provision the machine
	url := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows", a.APIServer, machineID)
	path := "/var/vcap/bosh/baremetal-cpi-agent-env.json"

	jsonStr := fmt.Sprintf("{\"name\":\"Graph.CF.CreateVM\",\"options\":{\"defaults\": {\"file\": \"%s\",\"path\": \"%s\", \"env\": \"%s\"}}}", stemcellCID, path, strings.Replace(string(jsonBytes), "\"", "\\\"", -1))
	a.logger.Info(a.logTag, "Json string is '%s'", jsonStr)

	resp, err := client.Post(url, strings.NewReader(jsonStr))
	if err != nil {
		bosherr.WrapErrorf(err, "Error uploading stemcell")
	}

	if (resp.StatusCode != http.StatusCreated) {
		return VMCID(""), errors.New("Error creating nodes")
	}

	//succeeded in creating workflow, so wait for the workflow to finish before returning
	for a.isWorkFlowActive(client, machineID) {
		time.Sleep(30 * time.Second)
	}

	//TODO implement full create vm using apis
	return VMCID(machineID), nil
}

func (a CreateVM) isWorkFlowActive(client httpclient.HTTPClient, machineID string) bool {
	a.logger.Info(a.logTag, "Checking workflow actively")
	workFlowUrl := fmt.Sprintf("http://%s:8080/api/common/nodes/%s/workflows/active", a.APIServer, machineID)
	resp, err := client.Get(workFlowUrl)

	if (err != nil) {
		//maybe better/diff error handling
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

type Workflow struct {
	Status *string `json:"_status"`
}