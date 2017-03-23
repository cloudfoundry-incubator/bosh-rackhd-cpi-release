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

//HasVM will check all nodes available to RACKHD for the given CID, true/false if CID exists.
func HasVM(c config.Cpi, extInput bosh.MethodArguments) (bool, error) {
	var cid string
	if reflect.TypeOf(extInput[0]) != reflect.TypeOf(cid) {
		return false, errors.New("Received unexpected type for vm cid")
	}

	cid = extInput[0].(string)

	nodes, err := rackhdapi.GetNodesByTag(c, cid)
	if err != nil {
		log.Info(fmt.Sprintf("Error found looking for vm cid %s. Info: %s", cid, err))
		return false, err
	}
	if len(nodes) > 1 {
		return false, fmt.Errorf("Returned %d too many nodes for 'HasVM'", (len(nodes) - 1))
	}
	// if there is one, return true! if not, false.
	return len(nodes) == 1, nil
}
