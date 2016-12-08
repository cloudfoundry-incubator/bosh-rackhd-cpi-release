package cpi

import (
  "encoding/json"
  "fmt"
  "reflect"

  "github.com/rackhd/rackhd-cpi/bosh"
  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/models"
  "github.com/rackhd/rackhd-cpi/rackhdapi"
)

func CreateDisk(c config.Cpi, extInput bosh.MethodArguments) (string, error) {
  diskSizeInMB, vmCID, err := parseCreateDiskInput(extInput)
  if err != nil {
    return "", err
  }

  filter := Filter{
    data:   diskSizeInMB,
    method: FilterBasedOnSizeMethod,
  }
  var diskCID string
  var node models.TagNode
  if vmCID != "" {
    node, err = rackhdapi.GetNodeByVMCID(c, vmCID)
    if err != nil {
      return "", err
    }

    if node.PersistentDisk.DiskCID != "" {
      return "", fmt.Errorf("error creating disk: VM %s already has a persistent disk", vmCID)
    }

    valid, err := filter.Run(c, node)
    if !valid || err != nil {
      return "", fmt.Errorf("error creating disk: %v", err)
    }

    if node.PersistentDisk.PregeneratedDiskCID == "" {
      return "", fmt.Errorf("error creating disk: can not find pregenerated disk cid for VM %s", vmCID)
    }
    diskCID = node.PersistentDisk.PregeneratedDiskCID

  } else {
    node.ID, err = TryReservationWithFilter(c, "", filter, SelectNodeFromRackHD, ReserveNodeFromRackHD)
    if err != nil {
      return "", err
    }
    diskCID = fmt.Sprintf("%s-%s", node.ID, c.RequestID)
  }

  container := models.PersistentDiskSettingsContainer{
    PersistentDisk: models.PersistentDiskSettings{
      DiskCID:    diskCID,
      Location:   fmt.Sprintf("/dev/%s", models.PersistentDiskLocation),
      IsAttached: false,
    },
  }

  bodyBytes, err := json.Marshal(container)
  if err != nil {
    return "", fmt.Errorf("error marshalling persistent disk information for VM %s", vmCID)
  }

  err = rackhdapi.PatchNode(c, node.ID, bodyBytes)
  if err != nil {
    return "", err
  }

  err = rackhdapi.CreateTag(c, node.ID, diskCID)
  if err != nil {
    return "", err
  }

  return container.PersistentDisk.DiskCID, nil
}

func parseCreateDiskInput(extInput bosh.MethodArguments) (int, string, error) {
  diskSizeInput := extInput[0]
  diskSizeInMB := int(diskSizeInput.(float64))

  vmCIDInput := extInput[2]
  var vmCID string
  if reflect.TypeOf(vmCID) != reflect.TypeOf(vmCIDInput) {
    return 0, "", fmt.Errorf("vmCIDInput is unexpected type")
  }

  vmCID = vmCIDInput.(string)

  return diskSizeInMB, vmCID, nil
}
