package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

const defaultMaxCreateVMAttempts = 5

func DefaultMaxCreateVMAttempts() int { return defaultMaxCreateVMAttempts }

func New(config io.Reader) (Cpi, error) {
	b, err := ioutil.ReadAll(config)
	if err != nil {
		log.Printf("Error reading config file %s", err)
		return Cpi{}, err
	}

	var cpi Cpi
	err = json.Unmarshal(b, &cpi)
	if err != nil {
		log.Printf("Error unmarshalling cpi config %s", err)
		return Cpi{}, err
	}

	if cpi.ApiServer == "" {
		log.Printf("ApiServer IP is not set")
		return Cpi{}, errors.New("ApiServer IP is not set")
	}

	if cpi.MaxCreateVMAttempt < 0 {
		log.Println("Invalid config. MaxCreateVMAttempt cannot be negative")
		return Cpi{}, errors.New("Invalid config. MaxCreateVMAttempt cannot be negative")
	}

	if cpi.MaxCreateVMAttempt == 0 {
		log.Printf("No MaxCreateVMAttempt was set, set to default value %d", defaultMaxCreateVMAttempts)
		cpi.MaxCreateVMAttempt = defaultMaxCreateVMAttempts
	}

	if !isAgentConfigValid(cpi.Agent) {
		log.Printf("Agent config invalid %v", cpi.Agent)
		return Cpi{}, fmt.Errorf("Agent config invalid %v", cpi.Agent)
	}

	return cpi, nil
}

func isAgentConfigValid(config AgentConfig) bool {
	if config.Mbus == "" {
		return false
	}

	// ntp is optional
	return true
}

type Cpi struct {
	ApiServer          string      `json:"apiserver"`
	Agent              AgentConfig `json:"agent"`
	MaxCreateVMAttempt int         `json:"max_create_vm_attempts"`
}

type AgentConfig struct {
	Mbus string `json:"mbus"`
	Ntp  string `json:"ntp"`
}
