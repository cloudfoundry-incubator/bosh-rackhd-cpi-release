package cpi

import (
  "errors"
  "reflect"

  "github.com/rackhd/rackhd-cpi/bosh"
  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/rackhdapi"
)

func HasVM(c config.Cpi, extInput bosh.MethodArguments) (bool, error) {
  var cid string
  if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
    return false, errors.New("Received unexpected type for vm cid")
  }

  cid = extInput[0].(string)

  _, err := rackhdapi.GetNodeByVMCID(c, cid)
  if err != nil {
    return false, err
  }

  return true, nil
}
