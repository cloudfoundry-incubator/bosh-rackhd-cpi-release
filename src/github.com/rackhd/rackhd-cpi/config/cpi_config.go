package config

import (
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "io/ioutil"
  "time"

  log "github.com/Sirupsen/logrus"
  "github.com/nu7hatch/gouuid"
  "github.com/rackhd/rackhd-cpi/bosh"
)

const (
  defaultMaxReserveNodeAttempts    = 5
  defaultRunWorkflowTimeoutSeconds = 20 * 60
)

type Cpi struct {
  ApiServer                 string        `json:"api_url"`
  Agent                     AgentConfig   `json:"agent"`
  MaxReserveNodeAttempts    int           `json:"max_reserve_node_attempts"`
  RunWorkflowTimeoutSeconds time.Duration `json:"run_workflow_timeout"`
  RequestID                 string        `json:"request_id"`
}

type AgentConfig struct {
  Blobstore map[string]interface{}
  Mbus      string   `json:"mbus"`
  Ntp       []string `json:"ntp"`
}

func DefaultMaxReserveNodeAttempts() int { return defaultMaxReserveNodeAttempts }

func GetNewRandomSeed() int64 { return time.Now().UnixNano() }

func New(config io.Reader, request bosh.CpiRequest) (Cpi, error) {
  b, err := ioutil.ReadAll(config)
  if err != nil {
    return Cpi{}, fmt.Errorf("Error reading config file %s", err)
  }

  var cpi Cpi
  err = json.Unmarshal(b, &cpi)
  if err != nil {
    return Cpi{}, fmt.Errorf("Error unmarshalling c config %s", err)
  }

  if cpi.ApiServer == "" {
    return Cpi{}, errors.New("ApiServer IP is not set")
  }

  if cpi.MaxReserveNodeAttempts < 0 {
    return Cpi{}, errors.New("Invalid config. MaxReserveNodeAttempts cannot be negative")
  }

  if cpi.MaxReserveNodeAttempts == 0 && (request.Method == bosh.CREATE_VM || request.Method == bosh.CREATE_DISK) {
    log.Info(fmt.Sprintf("No MaxReserveNodeAttempts was set, set to default value %d", defaultMaxReserveNodeAttempts))
    cpi.MaxReserveNodeAttempts = defaultMaxReserveNodeAttempts
  }

  if cpi.RunWorkflowTimeoutSeconds == 0 {
    log.Info(fmt.Sprintf("No RunWorkflowTimeoutSecounds was set, set to default value %d", defaultRunWorkflowTimeoutSeconds))
    cpi.RunWorkflowTimeoutSeconds = defaultRunWorkflowTimeoutSeconds
  }

  if cpi.RequestID == "" {
    uuid, err := uuid.NewV4()
    if err != nil {
      return Cpi{}, fmt.Errorf("Error generating uuid")
    }
    cpi.RequestID = uuid.String()
    log.Info(fmt.Sprintf("Using uuid for request: %s", cpi.RequestID))
  } else {
    log.Info(fmt.Sprintf("Using specified id for request: %s", cpi.RequestID))
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
