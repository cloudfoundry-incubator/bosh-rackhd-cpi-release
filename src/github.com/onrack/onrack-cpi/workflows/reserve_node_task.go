package workflows

import "github.com/onrack/onrack-cpi/onrackapi"

type reserveNodeTaskOptions struct {
	UUID     *string  `json:"uuid"`
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

var reserveNodeTaskTemplate = []byte(`{
  "friendlyName" : "Reserve Node",
  "injectableName" : "Task.BOSH.Reserve.Node",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "uuid": null,
    "commands":[
      "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"reserved\": \"{{ options.uuid }}\" }'"
    ]
  },
  "properties": {}
}`)
