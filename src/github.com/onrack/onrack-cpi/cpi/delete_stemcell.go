package cpi

import (
	"errors"
	"reflect"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
)

func DeleteStemcell(c config.Cpi, extInput bosh.MethodArguments) error {
	var cid string

	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
		return errors.New("Received unexpected type for stemcell cid")
	}

	cid = extInput[0].(string)

	return onrackapi.DeleteFile(c, cid)
}
