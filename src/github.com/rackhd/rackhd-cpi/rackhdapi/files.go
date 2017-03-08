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
)

type FileUploadResponse struct {
	Name   string `json:"name"`
	UUID   string `json:"uuid"`
	Md5    string `json:"md5"`
	Sha256 string `json:"sha256"`
}

func UploadFile(c config.Cpi, baseName string, r io.Reader, contentLength int64) (FileUploadResponse, error) {
	url := fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, baseName)
	request, err := http.NewRequest("PUT", url, ioutil.NopCloser(r))
	if err != nil {
		return FileUploadResponse{}, fmt.Errorf("Error building request to api server: %s", err)
	}
	request.ContentLength = contentLength

	respBody, err := helpers.MakeConfigRequest(request, []int{201})
	if err != nil {
		return FileUploadResponse{}, fmt.Errorf("Error making request %s", err)
	}

	var fileUploadResponse FileUploadResponse
	err = json.Unmarshal(respBody, &fileUploadResponse)
	if err != nil {
		return FileUploadResponse{}, fmt.Errorf("Error getting response from uploading file name: %s. error: %s", baseName, err)
	}

	return fileUploadResponse, nil
}

func DeleteFile(c config.Cpi, fileUUID string) error {
	url := fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, fileUUID)
	_, err := helpers.MakeRequestWithMultiCode(url, "DELETE", []int{204, 404}, nil)
	return err
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
