package models

const (
	OBMSettingIPMIServiceName = "ipmi-obm-service"
	OBMSettingAMTServiceName  = "amt-obm-service"
)

const (
	WorkflowSuccessfulStatus = "succeeded"
	WorkflowFailedStatus     = "failed"
	WorkflowCancelledStatus  = "cancelled"
	WorkflowRunningStatus    = "running"
	WorkflowPendingStatus    = "pending"
)

const (
	RackHDReserveVMGraphName = "Graph.CF.ReserveVM"
	RackHDCreateVMGraphName  = "Graph.BOSH.ProvisionNode"
	RackHDDeleteVMGraphName  = "Graph.CF.DeleteVM"
	RackHDEnvPath            = "/var/vcap/bosh/agent-bootstrap-env.json"
	DefaultUnusedName        = "UPLOADED_BY_RACKHD_CPI"
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

type Graph struct {
	Name       string                 `json:"injectableName"`
	UnusedName string                 `json:"friendlyName"`
	Options    map[string]interface{} `json:"options"`
	Tasks      []WorkflowTask         `json:"tasks"`
}

type WorkflowTask struct {
	TaskName string             `json:"taskName"`
	Label    string             `json:"label"`
	WaitOn   *map[string]string `json:"waitOn,omitempty"`
}

type WorkflowResponse struct {
	Name       string `json:"injectableName"`
	Status     string `json:"status"`
	InstanceID string `json:"instanceId"`
}

type PropertyContainer struct {
	Properties interface{} `json:"properties"`
}

type OptionContainer struct {
	Options interface{} `json:"options"`
}

type RunWorkflowRequestBody struct {
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}
