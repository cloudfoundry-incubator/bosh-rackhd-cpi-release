package cpi

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
)

func DeleteVM(c config.Cpi, extInput bosh.MethodArguments) error {
	var cid string
	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
		log.Println("Received unexpected type for vm cid")
		return errors.New("Received unexpected type for vm cid")
	}

	cid = extInput[0].(string)

	nodes, err := onrackhttp.GetNodes(c)
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

	workflowReq := onrackhttp.RunWorkflowRequestBody{
		Name: onrackhttp.OnrackDeleteVMGraphName,
		Options: map[string]interface{}{
			"defaults": nil,
		},
	}

	err = onrackhttp.RunWorkflow(c, nodeID, workflowReq)
	if err != nil {
		return fmt.Errorf("error reserving node %s", err)
	}

	return nil
}
