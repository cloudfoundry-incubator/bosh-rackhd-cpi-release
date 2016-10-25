package rackhdapi_test

import (
  "encoding/json"
  "fmt"
  "net/http"

  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/helpers"
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

  Describe("Get Node by Tag", func() {
    Context("when searching for an node with existing tag", func() {
      It("should find nodes with the tag", func() {
        fakeTag := "fake-vm-cid"
        expectedNodes := helpers.LoadTagNodes("../spec_assets/tag_node_with_cid.json")
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
        expectedNodes := helpers.LoadTagNodes("../spec_assets/tag_nodes.json")
        expectedNodesData, err := json.Marshal(expectedNodes)
        Expect(err).ToNot(HaveOccurred())

        url := fmt.Sprintf("/api/2.0/tags/%s/nodes", fakeTag)
        helpers.AddHandler(server, "GET", url, http.StatusOK, expectedNodesData)

        nodes, err := rackhdapi.GetNodesByTag(c, fakeTag)
        Expect(err).ToNot(HaveOccurred())
        Expect(len(nodes)).To(Equal(2))
      })
    })
  })
})
