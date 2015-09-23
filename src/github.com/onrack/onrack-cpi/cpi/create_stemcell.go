package cpi

import (
	"errors"
	"log"
	"os"
	"reflect"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/bosh"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"
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
		log.Println("Error obtaining stemcell file handle")
		return "", errors.New("Error obtaining stemcell file handle")
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

	_, err = onrackhttp.UploadFile(c, uuid.String(), stemcellFile, fileInfo.Size())
	if err != nil {
		return "", err
	}

	return uuid.String(), nil
}
