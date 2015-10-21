package workflows

import "github.com/onrack/onrack-cpi/onrackapi"

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
	*onrackapi.TaskStub
	*onrackapi.PropertyContainer
	*reserveNodeTaskOptionsContainer
}

type reserveNodeTaskOptionsContainer struct {
	Options reserveNodeTaskOptions `json:"options"`
}
