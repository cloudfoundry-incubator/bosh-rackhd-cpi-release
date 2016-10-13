package rackhdapi

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/models"
)

func UploadFile(c config.Cpi, baseName string, r io.Reader, contentLength int64) (string, error) {
	url := fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, baseName)
	body := ioutil.NopCloser(r)
	request, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return "", fmt.Errorf("Error building request to api server: %s", err)
	}
	request.ContentLength = contentLength

	log.Debug(fmt.Sprintf("uploading file: %s to server", baseName))
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("Error making request to api server: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("Failed uploading %s with status: %s\nresponse: %+v", baseName, resp.Status, resp)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to read response body"))
		return "", err
	}

	log.Debug(fmt.Sprintf("uploaded file: %s to server", baseName))

	return string(bodyBytes), nil
}

func DeleteFile(c config.Cpi, baseName string) error {
	url := fmt.Sprintf("%s/api/2.0/files/%s/metadata", c.ApiServer, baseName)
	metadataResp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error getting file metadata: %s", err)
	}
	defer metadataResp.Body.Close()

	if metadataResp.StatusCode == 404 {
		log.Error(fmt.Sprintf("File with basename: %s has already been deleted", baseName))
		return nil
	}

	metadataBytes, err := ioutil.ReadAll(metadataResp.Body)
	if err != nil {
		return fmt.Errorf("error reading metadata response body %s", err)
	}

	metadata := models.FileMetadataResponse{}
	err = json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return fmt.Errorf("error unmarshalling metadata response: %s", err)
	}

	url = fmt.Sprintf("%s/api/2.0/files/%s", c.ApiServer, metadata.UUID)
	deleteReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request %s", err)
	}

	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		return fmt.Errorf("error performing delete request %s", err)
	}

	if deleteResp.StatusCode == 404 {
		log.Error(fmt.Sprintf("File with basename: %s has already been deleted", baseName))
		return nil
	}

	if deleteResp.StatusCode != 204 {
		return fmt.Errorf("Failed deleting: %s with status: %s", baseName, deleteResp.Status)
	}

	return nil
}
