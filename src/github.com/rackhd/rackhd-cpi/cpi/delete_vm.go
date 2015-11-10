package cpi

import (
	"errors"
	"reflect"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
	"github.com/rackhd/rackhd-cpi/workflows"
)

func DeleteVM(c config.Cpi, extInput bosh.MethodArguments) error {
	var cid string
	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
		return errors.New("Received unexpected type for vm cid")
	}

	cid = extInput[0].(string)

	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return err
	}

	var nodeID string
	for _, node := range nodes {
		if node.CID == cid {
			nodeID = node.ID
		}
	}

	if nodeID == "" {
		return errors.New("cid was not found")
	}

	workflowName, err := workflows.PublishDeprovisionNodeWorkflow(c, cid)
	if err != nil {
		return err
	}

	defer rackhdapi.ReleaseNode(c, nodeID)

	return workflows.RunDeprovisionNodeWorkflow(c, nodeID, workflowName)
}
