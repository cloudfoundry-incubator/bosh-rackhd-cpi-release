package cpi

import (
  "errors"
  "reflect"

  "github.com/rackhd/rackhd-cpi/bosh"
  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/rackhdapi"
)

func DeleteStemcell(c config.Cpi, extInput bosh.MethodArguments) error {
  var cid string

  if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
    return errors.New("Received unexpected type for stemcell cid")
  }

  cid = extInput[0].(string)

  return rackhdapi.DeleteFile(c, cid)
}
