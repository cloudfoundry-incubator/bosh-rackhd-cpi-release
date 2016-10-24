package helpers

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "net/http"
  "regexp"
  "strings"

  "github.com/nu7hatch/gouuid"
)

// MakeRequestWithMultiCode builds a request with given info and makes the request
func MakeRequestWithMultiCode(url, method string, statusCode []int, body []byte) ([]byte, error) {
  errMsg := fmt.Sprintf("%s request to %s with body %+v", method, url, string(body))

  req, err := http.NewRequest(method, url, bytes.NewReader(body))
  if err != nil {
    return nil, fmt.Errorf("error building %s: %s", errMsg, err)
  }
  req.Header.Add("Content-Type", "application/json")

  return MakeConfigedRequest(req, statusCode)
}

// MakeConfigedRequest makes the given request
func MakeConfigedRequest(req *http.Request, statusCode []int) ([]byte, error) {
  errMsg := fmt.Sprintf("%s request to %s with body %+v", req.Method, req.URL, req.Body)

  resp, err := http.DefaultClient.Do(req)
  defer resp.Body.Close()
  if err != nil {
    return nil, fmt.Errorf("error making %s: %s", errMsg, err)
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, fmt.Errorf("error parsing response body %s: %s", errMsg, err)
  }

  for _, code := range statusCode {
    if resp.StatusCode == code {
      return respBody, nil
    }
  }

  return nil, fmt.Errorf("error getting response from %s: %d, %s", errMsg, resp.StatusCode, string(respBody))
}

// MakeRequest builds a request by given info and make the request
func MakeRequest(url, method string, statusCode int, body []byte) ([]byte, error) {
  return MakeRequestWithMultiCode(url, method, []int{statusCode}, body)
}

// GenerateUUID generates an uuid
func GenerateUUID() (string, error) {
  uuid, err := uuid.NewV4()
  if err != nil {
    return "", fmt.Errorf("error generating UUID: %s", err)
  }

  return uuid.String(), nil
}

// BytesToArray converts bytes to array of string
func BytesToArray(b []byte) []string {
  rg := regexp.MustCompile("[\"\\[\\]\"]")
  array := strings.Split(rg.ReplaceAllString(string(b), ""), ",")
  return array
}