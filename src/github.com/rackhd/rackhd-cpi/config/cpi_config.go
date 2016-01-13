package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	defaultMaxCreateVMAttempts       = 5
	defaultRunWorkflowTimeoutSeconds = 20 * 60
)

type Cpi struct {
	ApiServer                 string        `json:"apiserver"`
	Agent                     AgentConfig   `json:"agent"`
	MaxCreateVMAttempt        int           `json:"max_create_vm_attempts"`
	RunWorkflowTimeoutSeconds time.Duration `json:"run_workflow_timeout"`
}

type AgentConfig struct {
	Blobstore map[string]interface{}
	Mbus      string   `json:"mbus"`
	Ntp       []string `json:"ntp"`
}

func DefaultMaxCreateVMAttempts() int { return defaultMaxCreateVMAttempts }

func New(config io.Reader) (Cpi, error) {
	b, err := ioutil.ReadAll(config)
	if err != nil {
		return Cpi{}, fmt.Errorf("Error reading config file %s", err)
	}

	var cpi Cpi
	err = json.Unmarshal(b, &cpi)
	if err != nil {
		return Cpi{}, fmt.Errorf("Error unmarshalling cpi config %s", err)
	}

	if cpi.ApiServer == "" {
		return Cpi{}, errors.New("ApiServer IP is not set")
	}

	if cpi.MaxCreateVMAttempt < 0 {
		return Cpi{}, errors.New("Invalid config. MaxCreateVMAttempt cannot be negative")
	}

	if cpi.MaxCreateVMAttempt == 0 {
		log.Info(fmt.Sprintf("No MaxCreateVMAttempt was set, set to default value %d", defaultMaxCreateVMAttempts))
		cpi.MaxCreateVMAttempt = defaultMaxCreateVMAttempts
	}

	if cpi.RunWorkflowTimeoutSeconds == 0 {
		log.Info(fmt.Sprintf("No RunWorkflowTimeoutSecounds was set, set to default value %d", defaultRunWorkflowTimeoutSeconds))
		cpi.RunWorkflowTimeoutSeconds = defaultRunWorkflowTimeoutSeconds
	}

	if !isAgentConfigValid(cpi.Agent) {
		return Cpi{}, fmt.Errorf("Agent config invalid %v", cpi.Agent)
	}

	return cpi, nil
}

func isAgentConfigValid(config AgentConfig) bool {
	if config.Mbus == "" {
		return false
	}

	if len(config.Blobstore) == 0 {
		return false
	}

	_, providerExist := config.Blobstore["provider"]
	if !providerExist {
		return false
	}

	// ntp is optional
	return true
}
