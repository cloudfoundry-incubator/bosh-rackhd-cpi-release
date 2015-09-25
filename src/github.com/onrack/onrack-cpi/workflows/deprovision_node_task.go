package workflows

import "github.com/onrack/onrack-cpi/onrackhttp"

type deprovisionNodeTaskOptions struct {
	Type     string   `json:"type"`
	Commands []string `json:"commands"`
}

type deprovisionnodeTaskOptionsContainer struct {
	Options deprovisionNodeTaskOptions `json:"options"`
}

type deprovisionNodeTask struct {
	*onrackhttp.TaskStub
	*onrackhttp.PropertyContainer
	*deprovisionnodeTaskOptionsContainer
}

var deprovisionNodeTaskTemplate = []byte(`{
  "friendlyName": "Deprovision Node",
  "implementsTask": "Task.Base.Linux.Commands",
  "injectableName": "Task.BOSH.DeprovisionNode",
  "options": {
    "type": "quick",
    "commands": [
        "sudo dd if=/dev/zero of=/dev/sda bs=1M count=100",
        "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"cid\": \"\" }'"
    ]
  },
  "properties": {}
}`)
