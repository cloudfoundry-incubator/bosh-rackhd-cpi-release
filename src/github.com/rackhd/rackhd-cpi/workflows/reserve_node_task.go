package workflows

import "github.com/rackhd/rackhd-cpi/rackhdapi"

var reserveNodeTaskTemplate = []byte(`{
  "friendlyName" : "Reserve Node",
  "injectableName" : "Task.BOSH.Reserve.Node",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "commands":[
      "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"status\": \"reserved\" }'"
    ]
  },
  "properties": {}
}`)

type reserveNodeTaskOptions struct {
	Commands []string `json:"commands"`
}

type reserveNodeTask struct {
	*rackhdapi.TaskStub
	*rackhdapi.PropertyContainer
	*reserveNodeTaskOptionsContainer
}

type reserveNodeTaskOptionsContainer struct {
	Options reserveNodeTaskOptions `json:"options"`
}
