package onrackhttp

const (
	OnrackReserveVMGraphName = "Graph.CF.ReserveVM"
	OnrackCreateVMGraphName  = "Graph.CF.CreateVM"
	OnrackEnvPath            = "/var/vcap/bosh/onrack-cpi-agent-env.json"
	OnrackRegistryPath       = "/var/vcap/bosh/agent.json"
	DefaultUnusedName        = "UPLOADED_BY_ONRACK_CPI"
)

type NodeWorkflow struct {
	NodeID         string `json:"node"`
	InjectableName string `json:"injectableName"`
	Status         string `json:"_status"`
}

type Workflow struct {
	Name       string                 `json:"injectableName"`
	UnusedName string                 `json:"friendlyName"`
	Tasks      []WorkflowTask         `json:"tasks"`
	Options    map[string]interface{} `json:"options"`
}

type WorkflowStub struct {
	Name       string `json:"injectableName"`
	UnusedName string `json:"friendlyName"`
}

type WorkflowResponse struct {
	Name    string                  `json:"injectableName"`
	Tasks   map[string]TaskResponse `json:"tasks"`
	Options map[string]interface{}  `json:"options"`
	Status  string                  `json:"_status"`
}

type WorkflowTask struct {
	TaskName      string            `json:"taskName"`
	Label         string            `json:"label"`
	WaitOn        map[string]string `json:"waitOn,omitempty"`
	IgnoreFailure bool              `json:"ignoreFailure,omitempty"`
}

type Task struct {
	ImplementsTask string                 `json:"implementsTask,omitempty"`
	Name           string                 `json:"injectableName"`
	UnusedName     string                 `json:"friendlyName"`
	Options        map[string]interface{} `json:"options"`
	Properties     map[string]interface{} `json:"properties"`
}

type TaskStub struct {
	Name           string `json:"injectableName"`
	UnusedName     string `json:"friendlyName"`
	ImplementsTask string `json:"implementsTask,omitempty"`
}

type PropertyContainer struct {
	Properties interface{} `json:"properties"`
}

type OptionContainer struct {
	Options interface{} `json:"options"`
}

type TaskResponse struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type RunWorkflowRequestBody struct {
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}
