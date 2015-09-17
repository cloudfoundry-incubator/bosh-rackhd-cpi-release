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
