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
 	node, err := rackhdapi.GetNodeByVMCID(c, cid)
 	if err != nil {
 		return err
 	}

 	//if node.PersistentDisk.IsAttached {
 	//	err = rackhdapi.MakeDiskRequest(c, node, false)
 	//	if err != nil {
 	//		return err
 	//	}
 	//}

 	workflowName, err := workflows.PublishDeprovisionNodeWorkflow(c)
 	if err != nil {
 		return err
 	}

 	err = workflows.RunDeprovisionNodeWorkflow(c, node.ID, workflowName)
 	if err != nil {
 		return err
 	}

 	//if node.PersistentDisk.DiskCID == "" {
	err = rackhdapi.ReleaseNode(c, node.ID)
	if err != nil {
 		return err
 	}
 	//}

 	return nil
}
