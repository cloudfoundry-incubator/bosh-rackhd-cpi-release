package rackhdapi

import (
	"encoding/json"
	"fmt"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
)

// GetTags gets all tags on the given node
func GetTags(c config.Cpi, nodeID string) ([]string, error) {
	url := fmt.Sprintf("%s/api/2.0/nodes/%s/tags", c.ApiServer, nodeID)

	body, err := helpers.MakeRequest(url, "GET", 200, nil)
	if err != nil {
		return nil, err
	}

	var tags []string
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// DeleteTag deletes the given tag on the given node
func DeleteTag(c config.Cpi, nodeID, tag string) error {
	url := fmt.Sprintf("%s/api/2.0/nodes/%s/tags/%s", c.ApiServer, nodeID, tag)

	_, err := helpers.MakeRequest(url, "DELETE", 204, nil)
	return err
}

// CreateTag creates the given tag on the given node
func CreateTag(c config.Cpi, nodeID, tag string) error {
	tags := models.Tags{
		T: []string{tag},
	}
	body, err := json.Marshal(tags)
	if err != nil {
		return nil
	}

	url := fmt.Sprintf("%s/api/2.0/nodes/%s/tags", c.ApiServer, nodeID)
	_, err = helpers.MakeRequest(url, "PATCH", 200, body)
	return err
}

// GetNodesByTag returns all nodes that have the given tag
func GetNodesByTag(c config.Cpi, tag string) ([]models.TagNode, error) {
	url := fmt.Sprintf("%s/api/2.0/tags/%s/nodes", c.ApiServer, tag)
	respBody, err := helpers.MakeRequest(url, "GET", 200, nil)
	if err != nil {
		return nil, err
	}

	var nodes []models.TagNode
	err = json.Unmarshal(respBody, &nodes)
	return nodes, err
}

// GetNodeByTag returns the uniq node with given tag
func GetNodeByTag(c config.Cpi, tag string) (models.TagNode, error) {
	nodes, err := GetNodesByTag(c, tag)
	if err != nil {
		return models.TagNode{}, err
	}

	if len(nodes) > 1 {
		return models.TagNode{}, fmt.Errorf("error getting node by tag %s: more than one node returned", tag)
	} else if len(nodes) == 0 {
		return models.TagNode{}, fmt.Errorf("error getting node by tag %s: no node returned", tag)
	}
	return nodes[0], nil
}

// GetNodeByVMCID return the node with given Cloud ID
func GetNodeByVMCID(c config.Cpi, cid string) (models.TagNode, error) {
	return GetNodeByTag(c, cid)
}

// GetComputeNodesWithoutTags returns all available nodes that are not blocked or reserved
func GetComputeNodesWithoutTags(c config.Cpi, tags []string) ([]models.Node, error) {
	var result []models.Node

	dontReturnNodes := map[string]bool{}
	for _, tag := range tags {
		notTheseNodes, err := GetNodesByTag(c, tag)
		if err != nil {
			return nil, err
		}
		for _, node := range notTheseNodes {
			dontReturnNodes[node.ID] = true
		}
	}

	allNodes, err := GetNodesWithType(c, "compute")
	if err != nil {
		return nil, err
	}

	for _, n := range allNodes {
		if !dontReturnNodes[n.ID] {
			result = append(result, n)
		}
	}

	return result, nil
}

// ReleaseNode delete unavailable tag on the node
func ReleaseNode(c config.Cpi, nodeID string) error {
	return DeleteTag(c, nodeID, models.Unavailable)
}
