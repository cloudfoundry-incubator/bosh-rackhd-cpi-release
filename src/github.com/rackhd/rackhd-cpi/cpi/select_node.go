package cpi

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
)

func SelectNodeFromRackHD(c config.Cpi, diskCID string) (string, error) {
	if diskCID != "" {
		node, err := rackhdapi.GetNodeByDiskCID(c, diskCID)

		if err != nil {
			return "", err
		}

		log.Info(fmt.Sprintf("selected node %s", node.ID))
		return node.ID, nil
	}

	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return "", err
	}

	nodeID, err := randomSelectAvailableNode(c, nodes)
	if err != nil || nodeID == "" {
		return "", err
	}

	log.Info(fmt.Sprintf("selected node %s", nodeID))
	return nodeID, nil
}

func randomSelectAvailableNode(c config.Cpi, nodes []rackhdapi.Node) (string, error) {

	t := time.Now().UnixNano()
	log.Debug(fmt.Sprintf("Using random number seed (i.e., UnixNano time): %v", t))
	rand.Seed(t)
	shuffle := rand.Perm(len(nodes))
	log.Debug(fmt.Sprintf("Shuffled array: %v", shuffle))
	for i := range shuffle {
		log.Debug(fmt.Sprintf("Trying node: %v", shuffle[i]))
		node := nodes[shuffle[i]]
		if nodeIsAvailable(c, node) {
			log.Debug(fmt.Sprintf("node: %s is available", node.ID))
			return node.ID, nil
		}
	}

	return "", errors.New("all nodes have been reserved")
}

func nodeIsAvailable(c config.Cpi, n rackhdapi.Node) bool {
	log.Debug(fmt.Sprintf("Getting active workflow"))
	workflow, err := rackhdapi.GetActiveWorkflows(c, n.ID)
	if err != nil {
		log.Error(fmt.Sprintf("Error getting active workflow on node %s: %v\n", n.ID, err))
	}
	log.Debug(fmt.Sprintf("Getting OBM settings"))
	obmSettings, err := rackhdapi.GetOBMSettings(c, n.ID)
	if err != nil {
		log.Error(fmt.Sprintf("Error getting OBM settings on node %s: %v\n", n.ID, err))
	}

	return (n.Status == "" || n.Status == rackhdapi.Available) &&
		(n.CID == "") &&
		reflect.DeepEqual(workflow, rackhdapi.WorkflowResponse{}) &&
		(len(obmSettings) > 0) &&
		!hasPersistentDisk(n)
}

func hasPersistentDisk(n rackhdapi.Node) bool {
	return n.PersistentDisk.DiskCID != ""
}
