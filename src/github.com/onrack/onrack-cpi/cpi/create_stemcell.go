package cpi

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	log "github.com/Sirupsen/logrus"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"
)

func CreateStemcell(c config.Cpi, extInput bosh.MethodArguments) (string, error) {
	var imagePath string

	if reflect.TypeOf(extInput[0]) == reflect.TypeOf(imagePath) {
		imagePath = extInput[0].(string)
	} else {
		return "", errors.New("Received unexpected type for stemcell image path")
	}

	stemcellFile, err := os.Open(imagePath)
	if err != nil {
		log.Error(fmt.Sprintf("Error obtaining stemcell file handle %s", err))
		return "", fmt.Errorf("Error obtaining stemcell file handle %s", err)
	}

	defer stemcellFile.Close()

	fileInfo, err := stemcellFile.Stat()
	if err != nil {
		return "", err
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return "", errors.New("Error generating UUID")
	}

	_, err = onrackapi.UploadFile(c, uuid.String(), stemcellFile, fileInfo.Size())
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}
