package cpi

/*
	We hope that some version of the BOSH Director will provide a well defined
	input interface, namely a valid JSON map, however, at this time we have an
	 array with heterogeneous elements
 */
type ExternalInput []interface{}

type Config struct {
	ApiServer string `json:"apiserver"`
	Agent AgentConfig `json:"agent"`
}

type AgentConfig struct{
	Mbus string `json:"mbus"`
	Ntp string `json:"ntp"`
}
