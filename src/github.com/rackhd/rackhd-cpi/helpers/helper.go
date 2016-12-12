package helpers

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
  "regexp"
  "strings"

  "github.com/nu7hatch/gouuid"
  "github.com/rackhd/rackhd-cpi/models"
)

// MakeRequestWithMultiCode builds a request with given info and makes the request
func MakeRequestWithMultiCode(url, method string, statusCode []int, body []byte) ([]byte, error) {
  errMsg := fmt.Sprintf("%s request to %s with body %+v", method, url, string(body))

  req, err := http.NewRequest(method, url, bytes.NewReader(body))
  if err != nil {
    return nil, fmt.Errorf("error building %s: %s", errMsg, err)
  }
  if body != nil {
    req.Header.Add("Content-Type", "application/json")
    req.ContentLength = int64(len(body))
  }

  return MakeConfigRequest(req, statusCode)
}

// MakeConfigRequest makes the given request
func MakeConfigRequest(req *http.Request, statusCode []int) ([]byte, error) {
  errMsg := fmt.Sprintf("%s request to %s with body %+v", req.Method, req.URL, req.Body)

  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    return nil, fmt.Errorf("error making %s: %s", errMsg, err)
  }
  defer resp.Body.Close()

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

func ReadFile(filePath string) ([]byte, error) {
  reader, err := os.Open(filePath)
  defer reader.Close()
  if err != nil {
    return nil, err
  }

  return ioutil.ReadAll(reader)
}

func AddIDForTask(template string) (string, []byte, error) {
  taskBytes, err := ReadFile(template)
  if err != nil {
    return "", nil, err
  }

  task := models.Task{}
  err = json.Unmarshal(taskBytes, &task)

  uuid, err := GenerateUUID()
  if err != nil {
    return "", nil, err
  }
  task.Name += uuid

  taskBytes, err = json.Marshal(task)
  if err != nil {
    return "", nil, err
  }
  return task.Name, taskBytes, nil
}
