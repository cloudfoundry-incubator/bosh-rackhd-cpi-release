package onrackhttp

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/onrack/onrack-cpi/config"
)

func UploadFile(c config.Cpi, baseName string, r io.Reader, contentLength int64) (string, error) {
	url := fmt.Sprintf("http://%s:8080/api/common/files/%s", c.ApiServer, baseName)
	body := ioutil.NopCloser(r)
	request, err := http.NewRequest("PUT", url, body)
	if err != nil {
		log.Printf("Error building request to api server: %s", err)
		return "", err
	}
	request.ContentLength = contentLength
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("Error making request to api server: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Unable to read response body")
		return "", err
	}

	if resp.StatusCode != 201 {
		log.Printf("Failed uploading %s with status: %s", baseName, resp.Status)
		return "", fmt.Errorf("Failed uploading %s with status: %s", baseName, resp.Status)
	}

	return string(bodyBytes), nil
}

func DeleteFile(c config.Cpi, uuid string) error {
	url := fmt.Sprintf("http://%s:8080/api/common/files/%s", c.ApiServer, uuid)
	deleteReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating delete request %s", err)
	}

	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		return fmt.Errorf("error performing delete request %s", err)
	}

	if deleteResp.StatusCode != 204 {
		return fmt.Errorf("Failed deleting %s with status: %s", uuid, deleteResp.Status)
	}

	return nil
}
