package cpi

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

func CreateStemcell(c config.Cpi, extInput bosh.MethodArguments) (string, error) {
	var imagePath string

	if reflect.TypeOf(extInput[0]) == reflect.TypeOf(imagePath) {
		imagePath = extInput[0].(string)
	} else {
		return "", errors.New("received unexpected type for stemcell image path")
	}

	stemcellFile, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("error obtaining stemcell file handle %s", err)
	}
	defer stemcellFile.Close()

	fileInfo, err := stemcellFile.Stat()
	if err != nil {
		return "", fmt.Errorf("error getting file's stats: %s", err)
	}

	fileName, err := helpers.GenerateUUID()
	if err != nil {
		return "", err
	}

	uploadFile, err := rackhdapi.UploadFile(c, fileName, stemcellFile, fileInfo.Size())
	if err != nil {
		return "", err
	}
	log.Debug(fmt.Sprintf("uploaded stemcell: %s to server", fileName))

	return uploadFile.UUID, nil
}
