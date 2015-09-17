package cpi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/onrack/onrack-cpi/bosh"
)

func parseCreateVMInput(extInput bosh.ExternalInput) (string, string, map[string]bosh.Network, error) {
	networkSpecs := map[string]bosh.Network{}
	agentIDInput := extInput[0]
	var agentID string

	if reflect.TypeOf(agentIDInput) != reflect.TypeOf(agentID) {
		log.Printf("agent id has unexpected type: %s. Expecting a string", reflect.TypeOf(agentIDInput))
		return "", "", networkSpecs, fmt.Errorf("agent id has unexpected type: %s. Expecting a string", reflect.TypeOf(agentIDInput))
	}

	agentID = agentIDInput.(string)
	if agentID == "" {
		log.Println("agent id cannot be empty")
		return "", "", networkSpecs, errors.New("agent id cannot be empty")
	}

	stemcellIDInput := extInput[1]
	var stemcellID string

	if reflect.TypeOf(stemcellIDInput) != reflect.TypeOf(stemcellID) {
		log.Printf("stemcell id has unexpected type: %s. Expecting a string", reflect.TypeOf(stemcellIDInput))
		return "", "", networkSpecs, fmt.Errorf("stemcell id has unexpected type: %s. Expecting a string", reflect.TypeOf(stemcellIDInput))
	}

	stemcellID = stemcellIDInput.(string)
	if stemcellID == "" {
		log.Println("stemcell id cannot be empty")
		return "", "", networkSpecs, errors.New("stemcell id cannot be empty")
	}

	networkInput := extInput[3]
	var networks map[string]interface{}

	if reflect.TypeOf(networkInput) != reflect.TypeOf(networks) {
		log.Printf("network config has unexpected type: %s. Expecting a map", reflect.TypeOf(networkInput))
		return "", "", networkSpecs, fmt.Errorf("network config has unexpected type in: %s. Expecting a map", reflect.TypeOf(networkInput))
	}

	networks = networkInput.(map[string]interface{})
	if len(networks) > 1 {
		log.Printf("config error: Only one network supported %d", len(networks))
		return "", "", networkSpecs, fmt.Errorf("config error: Only one network supported, provided length: %d", len(networks))
	}

	b, err := json.Marshal(networks)
	if err != nil {
		log.Printf("error marshalling the network")
		return "", "", networkSpecs, errors.New("error marshalling the network")
	}

	var boshNetworks map[string]bosh.Network
	err = json.Unmarshal(b, &boshNetworks)
	if err != nil {
		log.Printf("error unmarshalling the network")
		return "", "", networkSpecs, errors.New("error unmarshalling the network")
	}

	var boshNetName string
	var boshNet bosh.Network

	for k, v := range boshNetworks {
		boshNetName = k
		boshNet = v
	}

	if valErr := validateNetworkingConfig(boshNet); valErr != nil {
		return "", "", networkSpecs, valErr
	}

	networkSpecs[boshNetName] = bosh.Network{
		NetworkType: boshNet.NetworkType,
		Netmask:     boshNet.Netmask,
		Gateway:     boshNet.Gateway,
		IP:          boshNet.IP,
		Default:     boshNet.Default,
		DNS:         boshNet.DNS,
	}

	return agentID, stemcellID, networkSpecs, nil
}

func validateNetworkingConfig(bn bosh.Network) error {
	if bn.NetworkType == bosh.ManualNetworkType {
		return validateManualNetworkingConfig(bn)
	}

	if bn.NetworkType == bosh.DynamicNetworkType {
		return nil
	}

	return fmt.Errorf("unexpected or empty network type: %s\n", bn.NetworkType)
}

func validateManualNetworkingConfig(bn bosh.Network) error {
	if bn.IP == "" {
		log.Println("config error: ip must be specified for manual network")
		return errors.New("config error: ip must be specified for manual network")
	}

	if bn.Gateway == "" {
		log.Println("config error: gateway must be specified for manual network")
		return errors.New("config error: gateway must be specified for manual network")
	}

	if bn.Netmask == "" {
		log.Println("config error: netmask must be specified for manual network")
		return errors.New("config error: netmask must be specified for manual network")
	}

	return nil
}
