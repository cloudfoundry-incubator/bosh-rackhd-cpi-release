package cpi

import (
	"errors"
	"reflect"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
)

func HasVM(c config.Cpi, extInput bosh.MethodArguments) (bool, error) {
	var cid string
	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
		return false, errors.New("Received unexpected type for vm cid")
	}

	cid = extInput[0].(string)

	nodes, err := onrackapi.GetNodes(c)
	if err != nil {
		return false, err
	}

	for _, node := range nodes {
		if node.CID == cid {
			return true, nil
		}
	}

	return false, nil
}
