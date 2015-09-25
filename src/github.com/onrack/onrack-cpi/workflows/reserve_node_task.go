package workflows

import "github.com/onrack/onrack-cpi/onrackhttp"

type UploadAgentSettingsOptions struct {
	AgentSettingsFile    string `json:"agentSettingsFile"`
	AgentSettingsPath    string `json:"agentSettingsPath"`
	CID                  string `json:"cid"`
	DownloadDir          string `json:"downloadDir,omitempty"`
	PublicKeyFile        string `json:"publicKeyFile"`
	RegistrySettingsFile string `json:"registrySettingsFile"`
	RegistrySettingsPath string `json:"registrySettingsPath"`
	StemcellFile         string `json:"stemcellFile"`
}

type ReserveVMOptions struct {
	UUID string `json:"uuid"`
}

type SetIdAndRebootOptions struct {
	CID      string   `json:"cid"`
	Commands []string `json:"commands"`
}

type ReserveNodeOptions struct {
	UUID     *string  `json:"uuid"`
	Commands []string `json:"commands"`
}

type ReserveNodeTask struct {
	*onrackhttp.TaskStub
	*onrackhttp.PropertyContainer
	*ReserveNodeOptionsContainer
}

type ReserveNodeOptionsContainer struct {
	Options ReserveNodeOptions `json:"options"`
}

var reserveNodeTemplate = []byte(`{
  "friendlyName" : "Reserve Node",
  "injectableName" : "Task.Os.Reserve.CF.VM",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "uuid": null,
    "commands":[
      "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"reserved\": \"{{ options.uuid }}\" }'"
    ]
  },
  "properties": {}
}`)
