package cpi

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
	"github.com/rackhd/rackhd-cpi/workflows"
)

const (
	AllowAnyNodeMethod      = "AllowAnyNode"
	FilterBasedOnSizeMethod = "FilterBasedOnSize"
)

type selectionFunc func(config.Cpi, string, Filter) (rackhdapi.Node, error)
type reservationFunc func(config.Cpi, rackhdapi.Node) error

type Filter struct {
	data   interface{}
	method string
}

func TryReservation(c config.Cpi, nodeID string, choose selectionFunc, reserve reservationFunc) (string, error) {
	return TryReservationWithFilter(c, nodeID, Filter{nil, AllowAnyNodeMethod}, choose, reserve)
}

func TryReservationWithFilter(c config.Cpi, nodeID string, filter Filter, choose selectionFunc, reserve reservationFunc) (string, error) {
	var node rackhdapi.Node
	var err error
	for i := 0; i < c.MaxReserveNodeAttempts; i++ {
		node, err = choose(c, nodeID, filter)
		nodeID = node.ID
		if err != nil {
			log.Error(fmt.Sprintf("retry %d: error choosing node %s", i, err))
			continue
		}

		err = reserve(c, node)
		if err != nil {
			log.Error(fmt.Sprintf("retry %d: error reserving node %s", i, err))
			if strings.HasPrefix(err.Error(), "Timed out running workflow") {
				rackhdapi.ReleaseNode(c, node.ID)
			}
			rand.Seed(time.Now().UnixNano())
			sleepTime := rand.Intn(5000)
			log.Debug(fmt.Sprintf("Sleeping for %d ms\n", sleepTime))
			time.Sleep(time.Millisecond * time.Duration(sleepTime))
			continue
		}

		break
	}

	if err != nil {
		return "", fmt.Errorf("unable to reserve node: %v", err)
	}

	return node.ID, nil
}

func (f Filter) AllowAnyNode() (bool, error) {
	return true, nil
}

func (f Filter) Run(c config.Cpi, node rackhdapi.Node) (bool, error) {
	if f.method == AllowAnyNodeMethod {
		return f.AllowAnyNode()
	}
	if f.method == FilterBasedOnSizeMethod {
		return f.FilterBasedOnSize(c, node)
	}
	return false, fmt.Errorf("error running filter: filter method not valid: %s", f.method)
}

func (f Filter) FilterBasedOnSize(c config.Cpi, node rackhdapi.Node) (bool, error) {
	size, ok := f.data.(int)
	if !ok {
		return false, fmt.Errorf("error converting disk size: disk size must be convertible to int")
	}

	catalog, err := rackhdapi.GetNodeCatalog(c, node.ID)
	if err != nil {
		return false, fmt.Errorf("error getting catalog of VM: %s", node.ID)
	}

	persistentDiskSize := catalog.Data.BlockDevices[rackhdapi.PersistentDiskLocation].Size
	if persistentDiskSize == "" {
		return false, fmt.Errorf("error creating disk for node %s: no disk found at %s", node.ID, rackhdapi.PersistentDiskLocation)
	}
	availableSpaceInKB, err := strconv.Atoi(persistentDiskSize)
	if err != nil {
		return false, fmt.Errorf("error creating disk for node %s: %v", node.ID, err)
	}

	if availableSpaceInKB < size*1024 {
		return false, fmt.Errorf("error creating disk with size %vMB for node %s: insufficient available disk space", size, node.ID)
	}

	return true, nil
}

func ReserveNodeFromRackHD(c config.Cpi, node rackhdapi.Node) error {
	if node.Status == rackhdapi.Reserved {
		return nil
	}

	workflowName, err := workflows.PublishReserveNodeWorkflow(c)
	if err != nil {
		return fmt.Errorf("error publishing reserve workflow: %s", err)
	}

	err = workflows.RunReserveNodeWorkflow(c, node.ID, workflowName)
	if err != nil {
		return fmt.Errorf("error running reserve workflow: %s", err)
	}

	log.Info(fmt.Sprintf("reserved node %s", node.ID))
	return nil
}

func SelectNodeFromRackHD(c config.Cpi, nodeID string, filter Filter) (rackhdapi.Node, error) {
	if nodeID != "" {
		node, err := rackhdapi.GetNode(c, nodeID)

		if err != nil {
			return rackhdapi.Node{}, err
		}

		log.Info(fmt.Sprintf("selected node %s", node.ID))
		return node, nil
	}

	nodes, err := rackhdapi.GetNodes(c)
	if err != nil {
		return rackhdapi.Node{}, err
	}

	node, err := randomSelectAvailableNode(c, nodes, filter)
	if err != nil || node.ID == "" {
		return rackhdapi.Node{}, err
	}

	log.Info(fmt.Sprintf("selected node %s", node.ID))
	return node, nil
}

func randomSelectAvailableNode(c config.Cpi, nodes []rackhdapi.Node, filter Filter) (rackhdapi.Node, error) {
	rand.Seed(time.Now().UnixNano())
	shuffle := rand.Perm(len(nodes))
	log.Debug(fmt.Sprintf("Accessing nodes randomly with pattern: %v", shuffle))

	for i := range shuffle {
		node := nodes[shuffle[i]]
		log.Debug(fmt.Sprintf("Trying node: %v", node.ID))
		if nodeIsAvailable(c, node, filter) {
			log.Debug(fmt.Sprintf("node %s is available", node.ID))
			return node, nil
		}
	}

	return rackhdapi.Node{}, errors.New("all nodes have been reserved")
}

func nodeIsAvailable(c config.Cpi, n rackhdapi.Node, filter Filter) bool {
	return hasAvailableState(n) &&
		hasNotBeenFiltered(c, n, filter) &&
		hasNoActiveWorkflow(c, n.ID) &&
		hasOBMSettings(c, n.ID) &&
		!hasPersistentDisk(n)
}

func hasAvailableState(n rackhdapi.Node) bool {
	return (n.Status == "" || n.Status == rackhdapi.Available) && (n.CID == "")
}

func hasNotBeenFiltered(c config.Cpi, n rackhdapi.Node, filter Filter) bool {
	log.Debug(fmt.Sprintf("Applying filter"))
	valid, err := filter.Run(c, n)
	if err != nil {
		log.Error(fmt.Sprintf("Error applying filter to node %s: %v\n", n.ID, err))
	}

	return valid
}

func hasPersistentDisk(n rackhdapi.Node) bool {
	return n.PersistentDisk.DiskCID != ""
}

func hasNoActiveWorkflow(c config.Cpi, nodeID string) bool {
	log.Debug(fmt.Sprintf("Getting active workflow"))
	workflow, err := rackhdapi.GetActiveWorkflows(c, nodeID)
	if err != nil {
		log.Error(fmt.Sprintf("Error getting active workflow on node %s: %v\n", nodeID, err))
	}
	return reflect.DeepEqual(workflow, rackhdapi.WorkflowResponse{})
}

func hasOBMSettings(c config.Cpi, nodeID string) bool {
	log.Debug(fmt.Sprintf("Getting OBM settings"))
	obmSettings, err := rackhdapi.GetOBMSettings(c, nodeID)
	if err != nil {
		log.Error(fmt.Sprintf("Error getting OBM settings on node %s: %v\n", nodeID, err))
	}
	return len(obmSettings) > 0
}
