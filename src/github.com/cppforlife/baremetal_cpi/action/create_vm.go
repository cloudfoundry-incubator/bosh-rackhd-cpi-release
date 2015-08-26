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
    //create agent and network env
	vmNetworks := networks.AsVMNetworks()
	vmEnv := bwcvm.Environment(env)
	agentEnv := bwcvm.NewAgentEnvForVM(agentID, "some-vm", vmNetworks, vmEnv, a.agentOptions)
	jsonBytes, err := json.Marshal(agentEnv)
	if err != nil {
		return "", bosherr.WrapError(err, "Marshalling agent env")
	}

    //call api to provision the machine
	url := fmt.Sprintf("post path", a.APIServer, "machine_id")
	path := "/var/vcap/bosh/baremetal-cpi-agent-env.json"

	jsonStr := fmt.Sprintf("body", stemcellCID, path, strings.Replace(string(jsonBytes), "\"", "\\\"", -1))
	a.logger.Info(a.logTag, "Json string is '%s'", jsonStr)

	resp, err := client.Post(url, strings.NewReader(jsonStr))
	if err != nil {
		bosherr.WrapErrorf(err, "Error uploading stemcell")
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	a.logger.Info(a.logTag, "Response: '%s' '%s'\n", resp.Status, string(responseBody))

	//TODO implement full create vm using apis
	return VMCID("some-vm"), nil
}
