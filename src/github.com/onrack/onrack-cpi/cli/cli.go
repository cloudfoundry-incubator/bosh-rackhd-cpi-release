package cli

import (
	"errors"
	"strings"
	"fmt"
	"encoding/json"

	"github.com/onrack/onrack-cpi/cpi"
)

func ParseCommand(rawInput []byte) (string, cpi.ExternalInput, error) {
	splitInput := strings.Split(string(rawInput), " ")

	implemented, err := cpi.ImplementsMethod(splitInput[0])
	if err != nil {
		return "", "", err
	}

	if !implemented {
		return "", "", errors.New(fmt.Sprintf("Method %s is not implemented", splitInput[0]))
	}

	extInput := cpi.ExternalInput{}
	err = json.Unmarshal([]byte(splitInput[1]), &extInput)
	if err != nil {
		return "", errors.New("Error parsing args")
	}


	return splitInput[0], extInput, nil
}
