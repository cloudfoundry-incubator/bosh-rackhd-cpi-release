package workflows

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
	FriendlyName   string  `json:"friendlyName"`
	InjectableName string  `json:"injectableName"`
	Tasks          []Task  `json:"tasks"`
	Options        Options `json:"options"`
}

type Task struct {
	TaskName      string            `json:"taskName"`
	Label         string            `json:"label"`
	WaitOn        map[string]string `json:"waitOn",omitempty`
	IgnoreFailure bool              `json:"ignoreFailure",omitempty`
}

type Defaults struct {
	Cid         string `json:"cid"`
	DownloadDir string `json:"downloadDir"`
	Env         string `json:"env"`
	File        string `json:"file"`
	Path        string `json:"path"`
}

type Options struct {
	BootstrapUbuntu map[string]string `json:"bootstrap-ubuntu"`
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
