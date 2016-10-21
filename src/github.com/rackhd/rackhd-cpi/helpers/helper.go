package helpers

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "net/http"

  "github.com/nu7hatch/gouuid"
)

func MakeRequest(url, method string, statusCode int, body []byte) ([]byte, error) {
  errMsg := fmt.Sprintf("%s request to %s with body %+v", method, url, string(body))

  req, err := http.NewRequest(method, url, bytes.NewReader(body))
  if err != nil {
    return nil, fmt.Errorf("error building %s: %s", errMsg, err)
  }

  resp, err := http.DefaultClient.Do(req)
  defer resp.Body.Close()
  if err != nil {
    return nil, fmt.Errorf("error making %s: %s", errMsg, err)
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, fmt.Errorf("error parsing response body %s: %s", errMsg, err)
  }

  if resp.StatusCode != statusCode {
    return nil, fmt.Errorf("error getting response from %s: %d, %s", errMsg, resp.StatusCode, string(respBody))
  }
  return respBody, nil
}

func GenerateUUID() (string, error) {
  uuid, err := uuid.NewV4()
  if err != nil {
    return "", fmt.Errorf("error generating UUID: %s", err)
  }

  return uuid.String(), nil
}
