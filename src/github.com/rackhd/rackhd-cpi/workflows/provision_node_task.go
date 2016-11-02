package workflows

import "github.com/rackhd/rackhd-cpi/models"

type provisionNodeOptions struct {
  AgentSettingsFile   *string  `json:"agentSettingsFile"`
  AgentSettingsMd5Uri string   `json:"agentSettingsMd5Uri"`
  AgentSettingsPath   *string  `json:"agentSettingsPath"`
  AgentSettingsURI    string   `json:"agentSettingsUri"`
  Commands            []string `json:"commands"`
  Device              string   `json:"device"`
  DownloadDir         string   `json:"downloadDir"`
  Persistent          string   `json:"persistent"`
  StemcellFileMd5Uri  string   `json:"stemcellFileMd5Uri"`
  StemcellFile        *string  `json:"stemcellFile"`
  StemcellURI         string   `json:"stemcellUri"`
  WipeDisk            string   `json:"wipeDisk"`
}

type provisionNodeTask struct {
  *models.Task
}

type provisionNodeOptionsContainer struct {
  Options provisionNodeOptions `json:"options"`
}
