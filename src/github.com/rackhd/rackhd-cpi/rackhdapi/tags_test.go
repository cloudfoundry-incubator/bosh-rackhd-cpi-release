package rackhdapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/models"
	"github.com/rackhd/rackhd-cpi/rackhdapi"


	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Tags", func() {
	var c config.Cpi
	var nodeID string
	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
		c = config.Cpi{ApiServer: server.URL()}
		nodeID = "fake-node-id"

	})

	Describe("GetTags", func() {
		Context("when there are tags attached to nodes", func() {
			It("should return all tags without error", func() {
				url := fmt.Sprintf("/api/2.0/nodes/%s/tags", nodeID)
				expectedData := []byte("[\"tag1\",\"tag2\"]")
				helpers.AddHandler(server, "GET", url, http.StatusOK, expectedData)

				tags, err := rackhdapi.GetTags(c, nodeID)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(tags)).To(Equal(2))
				Expect(tags).To(ContainElement("tag1"))
				Expect(tags).To(ContainElement("tag2"))
			})
		})
	})

	Describe("Delete Tag", func() {
		Context("when deleting an exsiting tag", func() {
			It("should delete the tag", func() {
				fakeTag := "existing-tag"
				url := fmt.Sprintf("/api/2.0/nodes/%s/tags/%s", nodeID, fakeTag)
				helpers.AddHandler(server, "DELETE", url, http.StatusNoContent, nil)

				err := rackhdapi.DeleteTag(c, nodeID, fakeTag)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Lifecycle", func() {
		var testNode models.Node
		BeforeEach(func() {
			rack_server_url := os.Getenv("RACKHD_API_URL")
			Expect(rack_server_url).ToNot(BeEmpty())

			c = config.Cpi{ApiServer: rack_server_url}
			nodes, err := rackhdapi.GetNodes(c)

			Expect(err).ToNot(HaveOccurred())
			Expect(len(nodes)).To(BeNumerically(">", 1))
			testNode, err = rackhdapi.GetNode(c, nodes[0].ID)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("When call put a tag name to a node", func() {
			It("should attach a tag to a node", func() {
				//Create tags and verify that they are attached to a node
				err := rackhdapi.CreateTag(c, testNode.ID, "test_tag")
				Expect(err).ToNot(HaveOccurred())
				tags, err := rackhdapi.GetTags(c, testNode.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(tags).To(ContainElement("test_tag"))

				//Delete created tags and verify that they are there
				err = rackhdapi.DeleteTag(c, testNode.ID, "test_tag")
				Expect(err).ToNot(HaveOccurred())
				tags, err = rackhdapi.GetTags(c, testNode.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(tags).ToNot(ContainElement("test_tag"))
			})
		})
	})

	Describe("Get Node by Tag", func() {
		Context("when searching for an node with existing tag", func() {
			It("should find nodes with the tag", func() {
				fakeTag := "fake-vm-cid"
				expectedNodes := helpers.LoadTagNodes("../spec_assets/tag_nodes_with_vm_cid.json")
				expectedNodesData, err := json.Marshal(expectedNodes)
				Expect(err).ToNot(HaveOccurred())

				url := fmt.Sprintf("/api/2.0/tags/%s/nodes", fakeTag)
				helpers.AddHandler(server, "GET", url, http.StatusOK, expectedNodesData)

				node, err := rackhdapi.GetNodeByTag(c, fakeTag)
				Expect(err).ToNot(HaveOccurred())
				Expect(node).To(Equal(expectedNodes[0]))
			})
		})

		Context("when searching node with non-existing tag", func() {
			It("should return err", func() {
				fakeTag := "non-existing-tag"
				url := fmt.Sprintf("/api/2.0/tags/%s/nodes", fakeTag)
				helpers.AddHandler(server, "GET", url, http.StatusOK, []byte("[]"))

				_, err := rackhdapi.GetNodeByTag(c, fakeTag)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Get Nodes by Tag", func() {
		Context("when searching for nodes with existing tag", func() {
			It("should find nodes with the tag", func() {
				fakeTag := "reserved"
				expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_reserved.json")
				url := fmt.Sprintf("/api/2.0/tags/%s/nodes", fakeTag)
				helpers.AddHandler(server, "GET", url, http.StatusOK, expectedNodesBytes)

				nodes, err := rackhdapi.GetNodesByTag(c, fakeTag)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(nodes)).To(Equal(1))
			})
		})
	})

	Describe("GetComputeNodesWithoutTags", func() {
		Context("when there are nodes without given tags", func() {
			It("should return nodes without error", func() {
				blockedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_blocked.json")
				url := fmt.Sprintf("/api/2.0/tags/%s/nodes", models.Blocked)
				helpers.AddHandler(server, "GET", url, 200, blockedNodesBytes)

				reservedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_reserved.json")
				url = fmt.Sprintf("/api/2.0/tags/%s/nodes", models.Unavailable)
				helpers.AddHandler(server, "GET", url, 200, reservedNodesBytes)

				computeNodes := helpers.LoadJSON("../spec_assets/nodes_all.json")
				helpers.AddHandlerWithParam(server, "GET", "/api/2.0/nodes", "type=compute", 200, computeNodes)

				nodes, err := rackhdapi.GetComputeNodesWithoutTags(c, []string{models.Blocked, models.Unavailable})
				Expect(err).ToNot(HaveOccurred())
				Expect(len(nodes)).To(Equal(1))
				availableNodeID := "57fb9fb03fcc55c807add402"
				Expect(nodes[0].ID).To(Equal(availableNodeID))
			})
		})
	})
})
