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
		log.Error(fmt.Sprintf("received unexpected type for stemcell image path"))
		return "", errors.New("received unexpected type for stemcell image path")
	}

	stemcellFile, err := os.Open(imagePath)
	if err != nil {
		log.Error(fmt.Sprintf("error obtaining stemcell file handle %s", err))
		return "", fmt.Errorf("error obtaining stemcell file handle %s", err)
	}

	defer stemcellFile.Close()

	fileInfo, err := stemcellFile.Stat()
	if err != nil {
		log.Error(fmt.Sprintf("error getting file's stats: %s", err))
		return "", fmt.Errorf("error getting file's stats: %s", err)
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		log.Error(fmt.Sprintf("error generating UUID: %s", err))
		return "", fmt.Errorf("error generating UUID: %s", err)
	}

	_, err = onrackapi.UploadFile(c, uuid.String(), stemcellFile, fileInfo.Size())
	if err != nil {
		return "", err
	}
	log.Debug(fmt.Sprintf("uploaded stemcell: %s to server", uuid.String()))

	return uuid.String(), nil
}
