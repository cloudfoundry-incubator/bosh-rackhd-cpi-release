package workflows

import "github.com/rackhd/rackhd-cpi/rackhdapi"

var provisionNodeTemplate = []byte(`{
  "friendlyName": "Provision Node",
  "implementsTask": "Task.Base.Linux.Commands",
  "injectableName": "Task.BOSH.Provision.Node",
  "options": {
    "agentSettingsFile": null,
    "agentSettingsMd5Uri": "/api/1.1/files/md5/{{ options.agentSettingsFile }}/latest",
    "agentSettingsPath": null,
    "agentSettingsUri": "/api/1.1/files/{{ options.agentSettingsFile }}/latest",
    "commands": [
      "if {{ options.wipeDisk }}; then sudo dd if=/dev/zero of={{ options.persistent }} bs=1M count=100; fi",
      "curl --retry 3 {{ server.apiServerURI }}{{ options.stemcellUri }} -o {{ options.downloadDir }}/{{ options.stemcellFile }}",
      "curl --retry 3 {{ server.apiServerURI }}{{ options.agentSettingsUri }} -o {{ options.downloadDir }}/{{ options.agentSettingsFile }}",
      "curl {{ server.apiServerURI }}{{ options.stemcellFileMd5Uri }} | tr -d '\"' > /opt/downloads/stemcellFileExpectedMd5",
      "curl {{ server.apiServerURI }}{{ options.agentSettingsMd5Uri }} | tr -d '\"' > /opt/downloads/agentSettingsExpectedMd5",
      "md5sum {{ options.downloadDir }}/{{ options.stemcellFile }} | cut -d' ' -f1 > /opt/downloads/stemcellFileCalculatedMd5",
      "md5sum {{ options.downloadDir }}/{{ options.agentSettingsFile }} | cut -d' ' -f1 > /opt/downloads/agentSettingsCalculatedMd5",
      "test $(cat /opt/downloads/stemcellFileCalculatedMd5) = $(cat /opt/downloads/stemcellFileExpectedMd5)",
      "test $(cat /opt/downloads/agentSettingsCalculatedMd5) = $(cat /opt/downloads/agentSettingsExpectedMd5)",
      "sudo umount {{ options.device }} || true",
      "sudo tar --to-stdout -xvf {{ options.downloadDir }}/{{ options.stemcellFile }} | sudo dd of={{ options.device }}",
      "sudo sfdisk -R {{ options.device }}",
      "sudo mount {{ options.device }}1 /mnt",
      "sudo dd if=/dev/zero of={{ options.device }}2 bs=1M count=100",
      "sudo dd if=/dev/zero of={{ options.device }}3 bs=1M count=100",
      "sudo cp {{ options.downloadDir }}/{{ options.agentSettingsFile }} /mnt/{{ options.agentSettingsPath }}",
      "sudo sync"
    ],
    "device": "/dev/sda",
    "downloadDir": "/opt/downloads",
    "persistent": "/dev/sdb",
    "stemcellFile": null,
    "stemcellFileMd5Uri": "/api/1.1/files/md5/{{ options.stemcellFile }}/latest",
    "stemcellUri": "/api/1.1/files/{{ options.stemcellFile }}/latest",
    "wipeDisk": "true"
  },
  "properties": {}
}`)

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
	*rackhdapi.TaskStub
	*rackhdapi.PropertyContainer
	*provisionNodeOptionsContainer
}

type provisionNodeOptionsContainer struct {
	Options provisionNodeOptions `json:"options"`
}
