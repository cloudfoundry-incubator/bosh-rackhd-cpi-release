package cpi

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
	"github.com/rackhd/rackhd-cpi/workflows"
)

const (
	AllowAnyNodeMethod      = "AllowAnyNode"
	FilterBasedOnSizeMethod = "FilterBasedOnSize"
)

type selectionFunc func(config.Cpi, string, Filter) (models.Node, error)
type reservationFunc func(config.Cpi, string) error

// Filter is a struct to hold "Data" and "method" as filtering arguments
type Filter struct {
	data   interface{}
	method string
}

// TryReservation will attempt to reserve a given node with a reservationFunc and selectionFunc
func TryReservation(c config.Cpi, nodeID string, choose selectionFunc, reserve reservationFunc) (string, error) {
	return TryReservationWithFilter(c, nodeID, Filter{nil, AllowAnyNodeMethod}, choose, reserve)
}

// TryReservationWithFilter is same as TryReservation, but with Filter
func TryReservationWithFilter(c config.Cpi, nodeID string, filter Filter, choose selectionFunc, reserve reservationFunc) (string, error) {
	var node models.Node
	var err error
	for i := 0; i < c.MaxReserveNodeAttempts; i++ {
		node, err = choose(c, nodeID, filter)
		if err != nil {
			log.Error(fmt.Sprintf("retry %d: error choosing node %s", i, err))
			continue
		}

		err = reserve(c, node.ID)
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

// AllowAnyNode returns a filter for all nodes
func (f Filter) AllowAnyNode() (bool, error) {
	return true, nil
}

// Run will run a filter against a node
func (f Filter) Run(c config.Cpi, node models.Node) (bool, error) {
	if f.method == AllowAnyNodeMethod {
		return f.AllowAnyNode()
	}
	if f.method == FilterBasedOnSizeMethod {
		return f.FilterBasedOnSize(c, node)
	}
	return false, fmt.Errorf("error running filter: filter method not valid: %s", f.method)
}

// FilterBasedOnSize will generate a filter based on the size on the node
func (f Filter) FilterBasedOnSize(c config.Cpi, node models.Node) (bool, error) {
	size, ok := f.data.(int)
	if !ok {
		return false, fmt.Errorf("error converting disk size: disk size must be convertible to int")
	}

	catalog, err := rackhdapi.GetNodeCatalog(c, node.ID)
	if err != nil {
		return false, fmt.Errorf("error getting catalog of VM: %s", node.ID)
	}

	persistentDiskSize := catalog.Data.BlockDevices[models.PersistentDiskLocation].Size
	if persistentDiskSize == "" {
		return false, fmt.Errorf("error creating disk for node %s: no disk found at %s", node.ID, models.PersistentDiskLocation)
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

// ReserveNodeFromRackHD will reserve a given node node from rackHD
func ReserveNodeFromRackHD(c config.Cpi, nodeID string) error {

	workflowName, err := workflows.PublishReserveNodeWorkflow(c)
	if err != nil {
		return fmt.Errorf("error publishing reserve workflow: %s", err)
	}

	err = workflows.RunReserveNodeWorkflow(c, nodeID, workflowName)
	if err != nil {
		return fmt.Errorf("error running reserve workflow: %s", err)
	}

	log.Info(fmt.Sprintf("reserved node %s", nodeID))
	return nil
}

// SelectNodeFromRackHD will try to select the node given, or return a random one
func SelectNodeFromRackHD(c config.Cpi, nodeID string, filter Filter) (models.Node, error) {
	if nodeID != "" {
		node, err := rackhdapi.GetNode(c, nodeID)

		if err != nil {
			return models.Node{}, err
		}

		log.Info(fmt.Sprintf("selected node %s", node.ID))
		return node, nil
	}

	nodes, err := rackhdapi.GetComputeNodesWithoutTags(c, []string{models.Reserved, models.Blocked, models.PersistentDisk})
	if err != nil {
		return models.Node{}, err
	}

	node, err := randomSelectNodeWithoutWorkflow(c, nodes, filter)
	if err != nil || node.ID == "" {
		return models.Node{}, err
	}

	log.Info(fmt.Sprintf("selected node %s", node.ID))
	return node, nil
}

func randomSelectNodeWithoutWorkflow(c config.Cpi, nodes []models.Node, filter Filter) (models.Node, error) {
	rand.Seed(time.Now().UnixNano())
	shuffle := rand.Perm(len(nodes))
	log.Debug(fmt.Sprintf("Accessing nodes randomly with pattern: %v", shuffle))

	for i := range shuffle {
		node := nodes[shuffle[i]]
		log.Debug(fmt.Sprintf("Trying node: %v", node.ID))

		if hasWorkflow, err := rackhdapi.HasActiveWorkflow(c, node.ID); err == nil {
			if !hasWorkflow {
				log.Debug(fmt.Sprintf("node %s is available", node.ID))
				return node, nil
			}
			continue
		} else {
			return models.Node{}, err
		}
	}

	return models.Node{}, errors.New("all nodes have been reserved")
}
