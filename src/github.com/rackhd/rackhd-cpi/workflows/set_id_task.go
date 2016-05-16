package workflows

import "github.com/rackhd/rackhd-cpi/rackhdapi"

var setNodeIDTemplate = []byte(`{
  "friendlyName" : "Set Id and Reboot VM",
  "injectableName" : "Task.BOSH.SetNodeId",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "cid": null,
    "commands":[
      "curl -X PATCH {{ server.apiServerURI }}/api/1.1/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"cid\": \"{{ options.cid }}\" }'"
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
	*rackhdapi.TaskStub
	*rackhdapi.PropertyContainer
	*setNodeIDOptionsContainer
}
