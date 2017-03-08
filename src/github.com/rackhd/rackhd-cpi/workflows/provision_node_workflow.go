package workflows

import (
	"encoding/json"
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

type ProvisionNodeWorkflowOptions struct {
	AgentSettingsFile    *string `json:"agentSettingsFile"`
	AgentSettingsPath    *string `json:"agentSettingsPath"`
	CID                  *string `json:"cid"`
	DownloadDir          string  `json:"downloadDir,omitempty"`
	OBMServiceName       *string `json:"obmServiceName"`
	RegistrySettingsFile *string `json:"registrySettingsFile"`
	RegistrySettingsPath *string `json:"registrySettingsPath"`
	StemcellFile         *string `json:"stemcellFile"`
	WipeDisk             string  `json:"wipeDisk"`
}

type provisionNodeWorkflowOptionsContainer struct {
	Options provisionNodeWorkflowDefaultOptionsContainer `json:"options"`
}

type provisionNodeWorkflowDefaultOptionsContainer struct {
	Defaults ProvisionNodeWorkflowOptions `json:"defaults"`
}

type provisionNodeWorkflow struct {
	*models.Graph
	*provisionNodeWorkflowOptionsContainer
	Tasks []models.WorkflowTask `json:"tasks"`
}

func RunProvisionNodeWorkflow(c config.Cpi, nodeID string, workflowName string, vmCID string, stemcellCID string, wipeDisk bool) error {
	options, err := buildProvisionWorkflowOptions(c, nodeID, vmCID, stemcellCID, wipeDisk)
	if err != nil {
		return err
	}

	req := models.RunWorkflowRequestBody{
		Name:    workflowName,
		Options: map[string]interface{}{"defaults": options},
	}

	return rackhdapi.RunWorkflow(rackhdapi.WorkflowPoster, rackhdapi.WorkflowFetcher, c, nodeID, req)
}

func PublishProvisionNodeWorkflow(c config.Cpi) (string, error) {
	tasks, workflow, err := generateProvisionNodeWorkflow(c.RequestID)
	if err != nil {
		return "", err
	}

	for i := range tasks {
		err = rackhdapi.PublishTask(c, tasks[i])
		if err != nil {
			return "", err
		}
	}

	w := provisionNodeWorkflow{}
	err = json.Unmarshal(workflow, &w)
	if err != nil {
		log.Error(fmt.Sprintf("error umarshalling workflow: %s", err))
		return "", err
	}

	err = rackhdapi.PublishGraph(c, workflow)
	if err != nil {
		return "", err
	}

	return w.Name, nil
}

func generateProvisionNodeWorkflow(uuid string) ([][]byte, []byte, error) {
	p := models.Task{}
	err := json.Unmarshal(provisionNodeTaskBytes, &p)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling provision node task template: %s\n", err))
		return nil, nil, fmt.Errorf("error unmarshalling provision node task template: %s\n", err)
	}

	p.Name = fmt.Sprintf("%s.%s", p.Name, uuid)
	p.UnusedName = fmt.Sprintf("%s.%s", p.UnusedName, "UPLOADED_BY_RACKHD_CPI")

	pBytes, err := json.Marshal(p)
	if err != nil {
		log.Error(fmt.Sprintf("error marshalling provision node task template: %s\n", err))
		return nil, nil, fmt.Errorf("error marshalling provision node task template: %s\n", err)
	}

	s := models.Task{}
	err = json.Unmarshal(setNodeIDTaskBytes, &s)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling set node id task template: %s\n", err))
		return nil, nil, fmt.Errorf("error unmarshalling set node id task template: %s\n", err)
	}

	s.Name = fmt.Sprintf("%s.%s", s.Name, uuid)
	s.UnusedName = fmt.Sprintf("%s.%s", s.UnusedName, "UPLOADED_BY_RACKHD_CPI")

	sBytes, err := json.Marshal(s)
	if err != nil {
		log.Error(fmt.Sprintf("error marshalling set node id task template: %s\n", err))
		return nil, nil, fmt.Errorf("error marshalling set node id task template: %s\n", err)
	}

	w := provisionNodeWorkflow{}
	err = json.Unmarshal(provisionNodeWorkflowBytes, &w)
	if err != nil {
		log.Error(fmt.Sprintf("error unmarshalling provision node workflow template: %s\n", err))
		return nil, nil, fmt.Errorf("error unmarshalling provision node workflow template: %s\n", err)
	}

	w.Name = fmt.Sprintf("%s.%s", w.Name, uuid)
	w.UnusedName = fmt.Sprintf("%s.%s", w.UnusedName, "UPLOADED_BY_RACKHD_CPI")
	w.Tasks[1].TaskName = fmt.Sprintf("%s.%s", w.Tasks[1].TaskName, uuid)
	w.Tasks[2].TaskName = fmt.Sprintf("%s.%s", w.Tasks[2].TaskName, uuid)

	wBytes, err := json.Marshal(w)
	if err != nil {
		log.Error(fmt.Sprintf("error marshalling provision node workflow template: %s\n", err))
		return nil, nil, fmt.Errorf("error marshalling provision node workflow template: %s\n", err)
	}

	return [][]byte{pBytes, sBytes}, wBytes, nil
}

func buildProvisionWorkflowOptions(c config.Cpi, nodeID string, vmCID string, stemcellCID string, wipeDisk bool) (ProvisionNodeWorkflowOptions, error) {
	envPath := models.RackHDEnvPath
	options := ProvisionNodeWorkflowOptions{
		AgentSettingsFile: &nodeID,
		AgentSettingsPath: &envPath,
		CID:               &vmCID,
		StemcellFile:      &stemcellCID,
		WipeDisk:          strconv.FormatBool(wipeDisk),
	}

	obmServiceName, err := rackhdapi.GetOBMServiceName(c, nodeID)
	if err != nil {
		return ProvisionNodeWorkflowOptions{}, fmt.Errorf("error retrieving obm settings from node: %s", nodeID)
	}
	options.OBMServiceName = &obmServiceName

	return options, nil
}

var provisionNodeTaskBytes = []byte(`
{
  "injectableName": "Task.BOSH.Node.Provision",
  "friendlyName": "Provision Node",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "agentSettingsFile": null,
    "agentSettingsMd5Uri": "{{ api.files }}/{{ options.agentSettingsFile }}/md5",
    "agentSettingsPath": null,
    "agentSettingsUri": "{{ api.files }}/{{ options.agentSettingsFile }}",
    "commands": [
      {
        "command": "if {{ options.wipeDisk }}; then sudo dd if=/dev/zero of={{ options.persistent }} bs=1M count=100; fi"
      },
      {
        "command": "curl --retry 3 {{ options.stemcellUri }} -o {{ options.downloadDir }}/{{ options.stemcellFile }}"
      },
      {
        "command": "curl --retry 3 {{ options.agentSettingsUri }} -o {{ options.downloadDir }}/{{ options.agentSettingsFile }}"
      },
      {
        "command": "curl {{ options.stemcellFileMd5Uri }} | tr -d '\"' > /opt/downloads/stemcellFileExpectedMd5"
      },
      {
        "command": "curl {{ options.agentSettingsMd5Uri }} | tr -d '\"' > /opt/downloads/agentSettingsExpectedMd5"
      },
      {
        "command": "md5sum {{ options.downloadDir }}/{{ options.stemcellFile }} | cut -d' ' -f1 > /opt/downloads/stemcellFileCalculatedMd5"
      },
      {
        "command": "md5sum {{ options.downloadDir }}/{{ options.agentSettingsFile }} | cut -d' ' -f1 > /opt/downloads/agentSettingsCalculatedMd5"
      },
      {
        "command": "test $(cat /opt/downloads/stemcellFileCalculatedMd5) = $(cat /opt/downloads/stemcellFileExpectedMd5)"
      },
      {
        "command": "test $(cat /opt/downloads/agentSettingsCalculatedMd5) = $(cat /opt/downloads/agentSettingsExpectedMd5)"
      },
      {
        "command": "sudo umount {{ options.device }} || true"
      },
      {
        "command": "sudo tar --to-stdout -xvf {{ options.downloadDir }}/{{ options.stemcellFile }} | sudo dd of={{ options.device }}"
      },
      {
        "command": "sudo sfdisk -R {{ options.device }}"
      },
      {
        "command": "sudo mount {{ options.device }}1 /mnt"
      },
      {
        "command": "sudo dd if=/dev/zero of={{ options.device }}2 bs=1M count=100"
      },
      {
        "command": "sudo dd if=/dev/zero of={{ options.device }}3 bs=1M count=100"
      },
      {
        "command": "sudo cp {{ options.downloadDir }}/{{ options.agentSettingsFile }} /mnt/{{ options.agentSettingsPath }}"
      },
      {
        "command": "sudo sync"
      }
    ],
    "device": "/dev/sda",
    "downloadDir": "/opt/downloads",
    "persistent": "/dev/sdb",
    "stemcellFile": null,
    "stemcellFileMd5Uri": "{{ api.files }}/{{ options.stemcellFile }}/md5",
    "stemcellUri": "{{ api.files }}/{{ options.stemcellFile }}",
    "wipeDisk": "true"
  },
  "properties": {}
}
`)

var provisionNodeWorkflowBytes = []byte(`
{
  "friendlyName": "BOSH Provision Node",
  "injectableName": "Graph.BOSH.Node.Provision",
  "options": {
    "defaults": {
      "agentSettingsFile": null,
      "agentSettingsPath": null,
      "cid": null,
      "downloadDir": "/opt/downloads",
      "obmServiceName": null,
      "registrySettingsFile": null,
      "registrySettingsPath": null,
      "stemcellFile": null,
      "wipeDisk": "true"
    }
  },
  "tasks": [
    {
      "label": "bootstrap-ubuntu",
      "taskName": "Task.Linux.Bootstrap.Ubuntu",
      "ignoreFailure": true
    },
    {
      "label": "provision-node",
      "taskName": "Task.BOSH.Node.Provision",
      "waitOn": {
        "bootstrap-ubuntu": "finished"
      }
    },
    {
      "label": "set-id",
      "taskName": "Task.BOSH.SetNodeId",
      "waitOn": {
        "provision-node": "succeeded"
      }
    },
    {
      "label": "reboot",
      "taskName": "Task.ProcShellReboot",
      "waitOn": {
        "set-id": "succeeded"
      }
    }
  ]
}
`)

var setNodeIDTaskBytes = []byte(`
{
  "friendlyName": "Set Id and Reboot VM",
  "injectableName": "Task.BOSH.SetNodeId",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "cid": null,
    "commands": [
      {
        "command": "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }}/tags -H \"Content-Type: application/json\" -d '{\"tags\": [\"{{ options.cid }}\"]}'"
      }
    ]
  },
  "properties": {}
}
`)
