package cpi_test

import (
  "encoding/json"
  "fmt"
  "net/http"
  "strings"

  "github.com/rackhd/rackhd-cpi/bosh"
  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/cpi"
  "github.com/rackhd/rackhd-cpi/helpers"
  "github.com/rackhd/rackhd-cpi/models"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "github.com/onsi/gomega/ghttp"
)

var _ = Describe("DeleteDisk", func() {
  var server *ghttp.Server
  var jsonReader *strings.Reader
  var cpiConfig config.Cpi
  var request bosh.CpiRequest

  BeforeEach(func() {
    server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.DELETE_DISK)
  })

  AfterEach(func() {
    server.Close()
  })

  Context("when given a disk cid for an existing, unattached disk", func() {
    var extInput bosh.MethodArguments
    var expectedDeleteDiskBodyBytes []byte
    var err error

    BeforeEach(func() {
      container := models.PersistentDiskSettingsContainer{
        PersistentDisk: models.PersistentDiskSettings{},
      }
      expectedDeleteDiskBodyBytes, err = json.Marshal(container)
      Expect(err).ToNot(HaveOccurred())
    })

    Context("when there is a VM left on the node", func() {
      It("deletes the disk and disk cid tag", func() {
        diskCID := "valid_disk_cid_3"
        jsonInput := []byte(`["` + diskCID + `"]`)
        err := json.Unmarshal(jsonInput, &extInput)
        Expect(err).NotTo(HaveOccurred())

        nodeID := "57fb9fb03fcc55c807add41c"
        expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_vm_disk_detached.json")
        server.AppendHandlers(
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", diskCID)),
            ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
          ),
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("PATCH", "/api/2.0/nodes/"+nodeID),
            ghttp.VerifyJSON(string(expectedDeleteDiskBodyBytes)),
          ),
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("DELETE", fmt.Sprintf("/api/2.0/nodes/%s/tags/%s", nodeID, diskCID)),
            ghttp.RespondWith(http.StatusNoContent, nil),
          ),
        )

        err = cpi.DeleteDisk(cpiConfig, extInput)
        Expect(len(server.ReceivedRequests())).To(Equal(3))
        Expect(err).NotTo(HaveOccurred())
      })
    })

    Context("when there is no VM left on the node", func() {
      It("deletes the disk and sets the status to available", func() {
        diskCID := "valid_disk_cid_1"
        jsonInput := []byte(`["` + diskCID + `"]`)
        err := json.Unmarshal(jsonInput, &extInput)
        Expect(err).NotTo(HaveOccurred())

        nodeID := "57fb9fb03fcc55c807add42b"
        expectedNodesBytes := helpers.LoadJSON("../spec_assets/tag_nodes_with_disk_cid.json")

        server.AppendHandlers(
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", diskCID)),
            ghttp.RespondWith(http.StatusOK, expectedNodesBytes),
          ),
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("PATCH", "/api/2.0/nodes/"+nodeID),
            ghttp.VerifyJSON(string(expectedDeleteDiskBodyBytes)),
            ghttp.RespondWith(http.StatusOK, nil),
          ),
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("DELETE", fmt.Sprintf("/api/2.0/nodes/%s/tags/%s", nodeID, diskCID)),
            ghttp.RespondWith(http.StatusNoContent, nil),
          ),
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("DELETE", fmt.Sprintf("/api/2.0/nodes/%s/tags/%s", nodeID, models.Unavailable)),
            ghttp.RespondWith(http.StatusNoContent, nil),
          ),
        )

        err = cpi.DeleteDisk(cpiConfig, extInput)
        Expect(err).NotTo(HaveOccurred())
        Expect(len(server.ReceivedRequests())).To(Equal(4))
      })
    })
  })

  Context("when given a disk cid for a non-existent disk", func() {
    It("returns an error", func() {
      diskCID := "invalid_disk_cid"
      jsonInput := []byte(`["` + diskCID + `"]`)
      var extInput bosh.MethodArguments
      err := json.Unmarshal(jsonInput, &extInput)
      Expect(err).ToNot(HaveOccurred())

      server.AppendHandlers(
        ghttp.CombineHandlers(
          ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", diskCID)),
          ghttp.RespondWith(http.StatusOK, []byte("[]")),
        ),
      )

      expectedErrorMsg := fmt.Sprintf("error getting node by tag %s: no node returned", diskCID)
      err = cpi.DeleteDisk(cpiConfig, extInput)
      Expect(err).To(MatchError(expectedErrorMsg))
      Expect(len(server.ReceivedRequests())).To(Equal(1))
    })
  })

  Context("when given a disk cid for a attached disk", func() {
    It("returns an error", func() {
      diskCID := "valid_disk_cid_2"
      jsonInput := []byte(`["` + diskCID + `"]`)
      var extInput bosh.MethodArguments
      err := json.Unmarshal(jsonInput, &extInput)
      Expect(err).ToNot(HaveOccurred())

      expectedNodes := helpers.LoadTagNodes("../spec_assets/tag_nodes_with_vm_disk_attached.json")
      expectedNodesData, err := json.Marshal(expectedNodes)
      Expect(err).ToNot(HaveOccurred())
      server.AppendHandlers(
        ghttp.CombineHandlers(
          ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/tags/%s/nodes", diskCID)),
          ghttp.RespondWith(http.StatusOK, expectedNodesData),
        ),
      )

      err = cpi.DeleteDisk(cpiConfig, extInput)
      Expect(err).To(MatchError("disk: " + diskCID + " is attached"))
      Expect(len(server.ReceivedRequests())).To(Equal(1))
    })
  })
})
