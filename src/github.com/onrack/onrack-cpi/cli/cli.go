package cli

import (
	"errors"
	"strings"
	"fmt"

	"github.com/onrack/onrack-cpi/cpi"
)

func ParseCommand(rawInput string) (string, string, error) {
	splitInput := strings.Split(rawInput, " ")

	implemented, err := cpi.ImplementsMethod(splitInput[0])
	if err != nil {
		return "", "", err
	}

	if !implemented {
		return "", "", errors.New(fmt.Sprintf("Method %s is not implemented", splitInput[0]))
	}

	return splitInput[0], splitInput[1], nil
}
