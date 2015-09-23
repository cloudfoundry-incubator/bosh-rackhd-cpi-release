package cpi

import (
	"errors"
	"log"
	"reflect"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
)

func DeleteStemcell(c config.Cpi, extInput bosh.MethodArguments) error {
	var cid string

	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
		log.Println("Received unexpected type for stemcell cid")
		return errors.New("Received unexpected type for stemcell cid")
	}

	cid = extInput[0].(string)

	return onrackhttp.DeleteFile(c, cid)
}
