package cpi

import (
	"errors"
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"
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
		log.Info(fmt.Sprintf("No node found for vm cid %s. Info: %s", cid, err))
		return false, nil
	}

	return true, nil
}
