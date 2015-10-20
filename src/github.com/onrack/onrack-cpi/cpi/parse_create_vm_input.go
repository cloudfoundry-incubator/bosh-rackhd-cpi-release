package cpi

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"

	"github.com/onrack/onrack-cpi/bosh"
)

func parseCreateVMInput(extInput bosh.MethodArguments) (string, string, string, map[string]bosh.Network, error) {
	networkSpecs := map[string]bosh.Network{}
	agentIDInput := extInput[0]
	var agentID string

	if reflect.TypeOf(agentIDInput) != reflect.TypeOf(agentID) {
		return "", "", "", networkSpecs, fmt.Errorf("agent id has unexpected type: %s. Expecting a string", reflect.TypeOf(agentIDInput))
	}

	agentID = agentIDInput.(string)
	if agentID == "" {
		return "", "", "", networkSpecs, errors.New("agent id cannot be empty")
	}

	stemcellIDInput := extInput[1]
	var stemcellID string

	if reflect.TypeOf(stemcellIDInput) != reflect.TypeOf(stemcellID) {
		return "", "", "", networkSpecs, fmt.Errorf("stemcell id has unexpected type: %s. Expecting a string", reflect.TypeOf(stemcellIDInput))
	}

	stemcellID = stemcellIDInput.(string)
	if stemcellID == "" {
		return "", "", "", networkSpecs, errors.New("stemcell id cannot be empty")
	}

	cloudPropertiesInput := extInput[2]
	var cloudProperties map[string]interface{}
	if reflect.TypeOf(cloudPropertiesInput) != reflect.TypeOf(cloudProperties) {
		return "", "", "", networkSpecs, fmt.Errorf("cloud properties has unexpected type: %s. Expecting a map to interface", reflect.TypeOf(cloudPropertiesInput))
	}

	cloudProperties = cloudPropertiesInput.(map[string]interface{})

	var encodedPublicKey string
	if publicKeyInput, keyExist := cloudProperties["public_key"]; keyExist {
		if reflect.TypeOf(publicKeyInput) != reflect.TypeOf(encodedPublicKey) {
			return "", "", "", networkSpecs, fmt.Errorf("public key has unexpected type: %s. Expecting a string", reflect.TypeOf(publicKeyInput))
		}
		encodedPublicKey = publicKeyInput.(string)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(encodedPublicKey)
	if err != nil {
		return "", "", "", networkSpecs, fmt.Errorf("unable to decode public key (base64) to string %s", err)
	}
	publicKey := string(publicKeyBytes)

	if publicKey == "" {
		log.Info("warning: public key is empty. You may not be able to log in to the machine")
	}

	networkInput := extInput[3]
	var networks map[string]interface{}

	if reflect.TypeOf(networkInput) != reflect.TypeOf(networks) {
		return "", "", "", networkSpecs, fmt.Errorf("network config has unexpected type in: %s. Expecting a map", reflect.TypeOf(networkInput))
	}

	networks = networkInput.(map[string]interface{})
	if len(networks) > 1 {
		return "", "", "", networkSpecs, fmt.Errorf("config error: Only one network supported, provided length: %d", len(networks))
	}

	b, err := json.Marshal(networks)
	if err != nil {
		return "", "", "", networkSpecs, errors.New("error marshalling the network")
	}

	var boshNetworks map[string]bosh.Network
	err = json.Unmarshal(b, &boshNetworks)
	if err != nil {
		return "", "", "", networkSpecs, errors.New("error unmarshalling the network")
	}

	var boshNetName string
	var boshNet bosh.Network

	for k, v := range boshNetworks {
		boshNetName = k
		boshNet = v
	}

	if valErr := validateNetworkingConfig(boshNet); valErr != nil {
		return "", "", "", networkSpecs, valErr
	}

	networkSpecs[boshNetName] = bosh.Network{
		NetworkType: boshNet.NetworkType,
		Netmask:     boshNet.Netmask,
		Gateway:     boshNet.Gateway,
		IP:          boshNet.IP,
		Default:     boshNet.Default,
		DNS:         boshNet.DNS,
	}

	return agentID, stemcellID, publicKey, networkSpecs, nil
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
		return errors.New("config error: ip must be specified for manual network")
	}

	if bn.Gateway == "" {
		return errors.New("config error: gateway must be specified for manual network")
	}

	if bn.Netmask == "" {
		return errors.New("config error: netmask must be specified for manual network")
	}

	return nil
}
