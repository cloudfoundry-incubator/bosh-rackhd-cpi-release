package workflows

import "github.com/onrack/onrack-cpi/onrackhttp"

type ReserveNodeOptions struct {
	UUID     *string  `json:"uuid"`
	Commands []string `json:"commands"`
}

type ReserveNodeTask struct {
	*onrackhttp.TaskStub
	*onrackhttp.PropertyContainer
	*ReserveNodeOptionsContainer
}

type ReserveNodeOptionsContainer struct {
	Options ReserveNodeOptions `json:"options"`
}

var reserveNodeTemplate = []byte(`{
  "friendlyName" : "Reserve Node",
  "injectableName" : "Task.Os.Reserve.CF.VM",
  "implementsTask": "Task.Base.Linux.Commands",
  "options": {
    "uuid": null,
    "commands":[
      "curl -X PATCH {{ api.base }}/nodes/{{ task.nodeId }} -H \"Content-Type: application/json\" -d '{\"reserved\": \"{{ options.uuid }}\" }'"
    ]
  },
  "properties": {}
}`)
