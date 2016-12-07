package cpi

import (
  "encoding/json"
  "fmt"
  "reflect"

  "github.com/rackhd/rackhd-cpi/bosh"
  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/rackhdapi"
)

func SetVMMetadata(c config.Cpi, extInput bosh.MethodArguments) error {
  var cid string
  if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
   return fmt.Errorf("Cannot set VM metadata: received unexpected value for vm cid: %s", cid)
  }
  cid = extInput[0].(string)

  metadata, err := json.Marshal(extInput[1])
  if err != nil {
    return fmt.Errorf("Cannot set VM metadata: metadata is not valid JSON")
  } else {
    for _, v := range extInput[1].(map[string]interface{}) {
      switch v.(type) {
      case string, int:
      default:
        fmt.Println("Cannot set VM metadata: metadata must be a hash with string or integer values")
      }
    }
  }

  node, err := rackhdapi.GetNodeByVMCID(c, cid)
  nodeID := node.ID
  if err != nil {
   return err
  }

  return rackhdapi.SetNodeMetadata(c, nodeID, string(metadata))
}
