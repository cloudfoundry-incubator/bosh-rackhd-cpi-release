package config

type Cpi struct {
	ApiServer string `json:"apiserver"`
	Agent AgentConfig `json:"agent"`
}

type AgentConfig struct{
	Mbus string `json:"mbus"`
	Ntp string `json:"ntp"`
}
