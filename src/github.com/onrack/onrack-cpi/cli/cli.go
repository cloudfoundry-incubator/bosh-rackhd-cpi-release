package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/cpi"
)

func ParseCommand(rawInput []byte) (string, bosh.ExternalInput, error) {
	splitInput := strings.Split(string(rawInput), " ")

	implemented, err := cpi.ImplementsMethod(splitInput[0])
	if err != nil {
		return "", bosh.ExternalInput{}, err
	}

	if !implemented {
		return "", bosh.ExternalInput{}, errors.New(fmt.Sprintf("Method %s is not implemented", splitInput[0]))
	}

	extInput := bosh.ExternalInput{}
	err = json.Unmarshal([]byte(splitInput[1]), &extInput)
	if err != nil {
		return "", bosh.ExternalInput{}, errors.New("Error parsing args")
	}

	return splitInput[0], extInput, nil
}
