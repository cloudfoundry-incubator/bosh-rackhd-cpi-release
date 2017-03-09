package rackhdapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
)

func UploadFile(c config.Cpi, baseName string, r io.Reader, contentLength int64) (models.FileUploadResponse, error) {
	url := fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, baseName)
	request, err := http.NewRequest("PUT", url, ioutil.NopCloser(r))
	if err != nil {
		return models.FileUploadResponse{}, fmt.Errorf("Error building request to api server: %s", err)
	}
	request.ContentLength = contentLength

	respBody, err := helpers.MakeConfigRequest(request, []int{201})
	if err != nil {
		return models.FileUploadResponse{}, fmt.Errorf("Error making request %s", err)
	}

	var fileUploadResponse models.FileUploadResponse
	err = json.Unmarshal(respBody, &fileUploadResponse)
	if err != nil {
		return models.FileUploadResponse{}, fmt.Errorf("Error getting response from uploading file name: %s. error: %s", baseName, err)
	}

	return fileUploadResponse, nil
}

func DeleteFile(c config.Cpi, fileName string) error {
	fileMetadata, err := getFileMetadata(c, fileName)
	if err != nil {
		return err
	}

	err = deleteFile(c, fileMetadata.UUID)
	if err != nil {
		return err
	}

	return nil
}

func GetFile(c config.Cpi, baseName string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, baseName)
	request, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		return []byte{}, fmt.Errorf("Error building request to api server: %s", err)
	}

	respBody, err := helpers.MakeConfigRequest(request, []int{200})
	if err != nil {
		return []byte{}, fmt.Errorf("Error making request %s", err)
	}

	return respBody, nil
}

func deleteFile(c config.Cpi, fileUUID string) error {
	url := fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, fileUUID)
	_, err := helpers.MakeRequestWithMultiCode(url, "DELETE", []int{204, 404}, nil)
	return err
}

func getFileMetadata(c config.Cpi, fileName string) (models.FileUploadResponse, error) {
	url := fmt.Sprintf("%s/api/2.0/files/%s/metadata", c.ApiServer, fileName)
	request, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		return models.FileUploadResponse{}, fmt.Errorf("Error building request to api server: %s", err)
	}

	respBody, err := helpers.MakeConfigRequest(request, []int{200})
	if err != nil {
		return models.FileUploadResponse{}, fmt.Errorf("Error making request %s", err)
	}

	var fileMetadata models.FileUploadResponse
	err = json.Unmarshal(respBody, &fileMetadata)
	if err != nil {
		return models.FileUploadResponse{}, fmt.Errorf("Error getting metadata from uploaded file name: %s. error: %s", fileName, err)
	}

	return fileMetadata, nil
}
