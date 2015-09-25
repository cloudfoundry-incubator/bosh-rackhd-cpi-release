package workflows

import "github.com/onrack/onrack-cpi/onrackhttp"

type setNodeIDThenRebootOptions struct {
	CID      *string  `json:"cid"`
	Commands []string `json:"commands"`
}

type setNodeIDThenRebootOptionsContainer struct {
	Options setNodeIDThenRebootOptions `json:"options"`
}

type setNodeIDThenRebootTask struct {
	*onrackhttp.TaskStub
	*onrackhttp.PropertyContainer
	*setNodeIDThenRebootOptionsContainer
}

var setNodeIDThenRebootTemplate = []byte(`{
  "friendlyName" : "Set Id and Reboot VM",
  "injectableName" : "Task.BOSH.SetNodeId",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "cid": null,
    "commands":[
      "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"cid\": \"{{ options.cid }}\" }'",
      "sudo reboot"
    ]
  },
  "properties": {}
}`)
