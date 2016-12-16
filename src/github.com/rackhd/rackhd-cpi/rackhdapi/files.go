package rackhdapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
)

func UploadFile(c config.Cpi, baseName string, r io.Reader, contentLength int64) (string, error) {
	url := fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, baseName)
	request, err := http.NewRequest("PUT", url, ioutil.NopCloser(r))
	if err != nil {
		return "", fmt.Errorf("Error building request to api server: %s", err)
	}
	request.ContentLength = contentLength

	respBody, err := helpers.MakeConfigRequest(request, []int{201})
	return string(respBody), nil
}

func DeleteFile(c config.Cpi, baseName string) error {
	url := fmt.Sprintf("%s/api/2.0/files/%s/metadata", c.ApiServer, baseName)
	respBody, err := helpers.MakeRequest(url, "GET", 200, nil)
	if err != nil {
		return err
	}

	metadata := models.FileMetadataResponse{}
	err = json.Unmarshal(respBody, &metadata)
	if err != nil {
		return fmt.Errorf("error unmarshalling metadata response: %s", err)
	}

	url = fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, metadata.UUID)
	_, err = helpers.MakeRequestWithMultiCode(url, "DELETE", []int{204, 404}, nil)
	return err
}
