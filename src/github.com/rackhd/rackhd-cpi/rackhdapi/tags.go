package rackhdapi

import (
  "bytes"
  "fmt"
  "io/ioutil"
  "net/http"
  "regexp"
  "strings"

  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/helpers"
)

func GetTags(c config.Cpi, nodeID string) ([]Tag, error) {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/tags", c.ApiServer, nodeID)

  body, err := helpers.MakeRequest(url, "GET", 200, nil)
  if err != nil {
    return nil, err
  }

  rg := regexp.MustCompile("[\\[\\]\"]")
  tagArray := strings.Split(rg.ReplaceAllString(string(body), ""), ",")
  tags := [len(tagArray)]Tag
  for , tag := range tagArray {

  }

  return tags, nil
}

func deleteTag(c config.Cpi, nodeID, tag string) error {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/tags/%s", c.ApiServer, nodeID, tag)

  _, err := helpers.MakeRequest(url, "DELETE", 200, nil)
  return err
}

func createTag(c config.Cpi, nodeID, tag string) error {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/tags", c.ApiServer, nodeID)

  _, err := helpers.MakeRequest(url, "PATCH", 200, nil)
  return err
}

func UpdateMetadata(c config.Cpi, nodeID, key, value string) error {
  url := fmt.Sprintf("%s/api/2.0/nodes/%s/tags", c.ApiServer, nodeID)
  body := []byte(fmt.Sprintf("{\"tags\": [\"%s\"]}", tag))


  return nil
}
