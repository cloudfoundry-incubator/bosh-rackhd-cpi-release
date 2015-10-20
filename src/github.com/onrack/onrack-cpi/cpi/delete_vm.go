package cpi

import (
	"errors"
	"reflect"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
	"github.com/onrack/onrack-cpi/workflows"
)

func DeleteVM(c config.Cpi, extInput bosh.MethodArguments) error {
	var cid string
	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
		return errors.New("Received unexpected type for vm cid")
	}

	cid = extInput[0].(string)

	nodes, err := onrackapi.GetNodes(c)
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

	return workflows.RunDeprovisionNodeWorkflow(c, nodeID, workflowName)
}
