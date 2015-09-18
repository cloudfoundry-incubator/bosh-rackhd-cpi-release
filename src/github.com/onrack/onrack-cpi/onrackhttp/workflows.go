package onrackhttp

const (
	OnrackReserveVMGraphName = "Graph.CF.CreateReserveVM"
	OnrackEnvPath            = "/var/vcap/bosh/onrack-cpi-agent-env.json"
	OnrackRegistryPath       = "/var/vcap/bosh/agent.json"
)

type NodeWorkflow struct {
	NodeID         string `json:"node"`
	InjectableName string `json:"injectableName"`
	Status         string `json:"_status"`
}

type Workflow struct {
	FriendlyName			string 									`json:"friendlyName"`
	InjectableName 		string 									`json:"injectableName"`
	Tasks 						[]WorkflowTask 					`json:"tasks"`
	Options						Options									`json:"options"`
}

type WorkflowTask struct {
	TaskName					string									`json:"taskName"`
	Label							string									`json:"label"`
	WaitOn						map[string]string				`json:"waitOn,omitempty"`
	IgnoreFailure			bool										`json:"ignoreFailure,omitempty"`
}

type Defaults struct {
  AgentSettingsFile     interface{}				      `json:"agentSettingsFile"`
  AgentSettingsPath     interface{}				      `json:"agentSettingsPath"`
  Cid                   interface{}				      `json:"cid"`
  DownloadDir           interface{}				      `json:"downloadDir,omitempty"`
  RegistrySettingsFile  interface{}				      `json:"registrySettingsFile"`
  RegistrySettingsPath  interface{}				      `json:"registrySettingsPath"`
  StemcellFile          interface{}				      `json:"stemcellFile"`
}

type Options struct {
	BootstrapUbuntu map[string]string `json:"bootstrap-ubuntu,omitempty"`
	Defaults        Defaults          `json:"defaults"`
}

type UploadAgentSettingsOptions struct {
	AgentSettingsFile    string `json:"agentSettingsFile"`
	AgentSettingsPath    string `json:"agentSettingsPath"`
	CID                  string `json:"cid"`
	DownloadDir          string `json:"downloadDir"`
	RegistrySettingsFile string `json:"registrySettingsFile"`
	RegistrySettingsPath string `json:"registrySettingsPath"`
	StemcellFile         string `json:"stemcellFile"`
}

type UploadAgentSettingsRequest struct {
	Name    string                                `json:"name"`
	Options map[string]UploadAgentSettingsOptions `json:"options"`
}

type Task struct {
	FriendlyName					   string									   `json:"friendlyName"`
	ImplementsTask				   string									   `json:"implementsTask,omitempty"`
	InjectableName				   string			               `json:"injectableName"`
	Options			             map[string]interface{}		 `json:"options"`
  Properties               map[string]interface{}    `json:"properties"`
}
