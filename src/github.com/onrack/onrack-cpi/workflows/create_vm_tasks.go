package workflows

type UploadAgentSettingsOptions struct {
	AgentSettingsFile    string `json:"agentSettingsFile"`
	AgentSettingsPath    string `json:"agentSettingsPath"`
	CID                  string `json:"cid"`
	DownloadDir          string `json:"downloadDir,omitempty"`
	PublicKeyFile        string `json:"publicKeyFile"`
	RegistrySettingsFile string `json:"registrySettingsFile"`
	RegistrySettingsPath string `json:"registrySettingsPath"`
	StemcellFile         string `json:"stemcellFile"`
}

type ReserveVMOptions struct {
	UUID string `json:"uuid"`
}

type SetIdAndRebootOptions struct {
	CID      string   `json:"cid"`
	Commands []string `json:"commands"`
}
