package workflows

import "github.com/onrack/onrack-cpi/onrackapi"

var setNodeIDTemplate = []byte(`{
  "friendlyName" : "Set Id and Reboot VM",
  "injectableName" : "Task.BOSH.SetNodeId",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "cid": null,
    "commands":[
      "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"cid\": \"{{ options.cid }}\" }'"
    ]
  },
  "properties": {}
}`)

type setNodeIDOptions struct {
	CID      *string  `json:"cid"`
	Commands []string `json:"commands"`
}

type setNodeIDOptionsContainer struct {
	Options setNodeIDOptions `json:"options"`
}

type setNodeIDTask struct {
	*onrackapi.TaskStub
	*onrackapi.PropertyContainer
	*setNodeIDOptionsContainer
}
