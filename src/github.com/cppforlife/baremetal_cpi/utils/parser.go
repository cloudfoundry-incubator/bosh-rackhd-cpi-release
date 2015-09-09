package utils

import(
	"io/ioutil"
	"net/http"
	"errors"
	"encoding/json"
)

func ReadResponseToJson(resp *http.Response, jsonObject interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Cannot read body")
	}

	err = json.Unmarshal(body, jsonObject)
	if err != nil {
		return errors.New("Cannot unmarshall the body")
	}

	return nil
}