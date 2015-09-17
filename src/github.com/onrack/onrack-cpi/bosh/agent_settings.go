package bosh

var defaultAgentRegistrySettings = AgentRegistrySettings{
	Infrastructure: agentInfrastructureSettings{
		Settings: agentRegistrySetting{
			Sources: []agentSettingsSource{
				agentSettingsSource{
					Settingspath: "/var/vcap/bosh/onrack-cpi-agent-env.json",
					Type:         "File",
				},
			},
			Useregistry: true,
		},
	},
}

func DefaultAgentRegistrySettings() AgentRegistrySettings {
	return defaultAgentRegistrySettings
}

const (
	DynamicNetworkType = "dynamic"
	ManualNetworkType  = "manual"
)

type Network struct {
	NetworkType     string                 `json:"type"`
	Netmask         string                 `json:"netmask"`
	Gateway         string                 `json:"gateway"`
	IP              string                 `json:"ip"`
	Default         []string               `json:"default"`
	DNS             []string               `json:"dns,omitempty"`
	CloudProperties map[string]interface{} `json:"cloud_properties"`
	MAC             string                 `json:"mac"`
}

type AgentEnv struct {
	AgentID   string                 `json:"agent_id"`
	Blobstore map[string]interface{} `json:"blobstore"`
	Disks     map[string]string      `json:"disks"`
	Env       map[string]interface{} `json:"env"`
	Mbus      string                 `json:"mbus"`
	Networks  map[string]Network     `json:"networks"`
	NTP       []string               `json:"ntp"`
	VM        map[string]string      `json:"vm"`
}

type AgentRegistrySettings struct {
	Infrastructure agentInfrastructureSettings `json:"Infrastructure"`
}

type agentInfrastructureSettings struct {
	Settings agentRegistrySetting `json:"Settings"`
}

type agentRegistrySetting struct {
	Sources     []agentSettingsSource `json:"Sources"`
	Useregistry bool                  `json:"UseRegistry"`
}

type agentSettingsSource struct {
	Settingspath string `json:"SettingsPath"`
	Type         string `json:"Type"`
}
