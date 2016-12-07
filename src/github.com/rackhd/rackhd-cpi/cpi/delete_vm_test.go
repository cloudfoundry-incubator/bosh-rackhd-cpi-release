package cpi_test

import (
 	"encoding/json"
 	"net/http"
 	"strings"

 	. "github.com/onsi/ginkgo"
 	. "github.com/onsi/gomega"

 	"github.com/onsi/gomega/ghttp"
 	"github.com/rackhd/rackhd-cpi/bosh"
 	"github.com/rackhd/rackhd-cpi/config"
 	"github.com/rackhd/rackhd-cpi/cpi"
 	"github.com/rackhd/rackhd-cpi/helpers"
  "github.com/rackhd/rackhd-cpi/models"
)

var _ = Describe("DeleteVM", func() {
 	var server *ghttp.Server
 	var jsonReader *strings.Reader
 	var cpiConfig config.Cpi
 	var request bosh.CpiRequest

 	BeforeEach(func() {
 		server, jsonReader, cpiConfig, request = helpers.SetUp(bosh.DELETE_VM)
 		cpiConfig.RequestID = "requestid"
 	})

 	AfterEach(func() {
 		server.Close()
 	})

 	Context("with a valid VM CID and valid states", func() {
 		var extInput bosh.MethodArguments

 		//Context("when there is a persistent disk left before deprovisioning", func() {
 		//	It("deprovisions the node", func() {
      //  expectedNodesData := helpers.LoadJSON("../spec_assets/tag_node_with_cid.json")
      //  server.AppendHandlers(
      //    ghttp.CombineHandlers(
      //      ghttp.VerifyRequest("GET", "/api/2.0/tags/vm-5678/nodes"),
      //      ghttp.RespondWith(http.StatusOK, expectedNodesData),
      //    ),
      //  )
     //
     //
 		//		jsonInput := []byte(`["vm-5678"]`)
 		//		err := json.Unmarshal(jsonInput, &extInput)
 		//		Expect(err).NotTo(HaveOccurred())
     //
 		//		server.AppendHandlers(
 		//			helpers.MakeWorkflowHandlers(
 		//				"Deprovision",
 		//				cpiConfig.RequestID,
 		//				"57fb9fb03fcc55c807add41c",
 		//			)...,
 		//		)
     //
 		//		err = cpi.DeleteVM(cpiConfig, extInput)
 		//		Expect(err).NotTo(HaveOccurred())
 		//		Expect(len(server.ReceivedRequests())).To(Equal(8))
 		//	})
 		//})

 		Context("when there is no persistent disk left before deprovisioning", func() {
 			It("deprovisions the node and sets the status to available", func() {
        expectedNodesData := helpers.LoadJSON("../spec_assets/tag_node_with_cid.json")
        server.AppendHandlers(
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", "/api/2.0/tags/vm-1234/nodes"),
            ghttp.RespondWith(http.StatusOK, expectedNodesData),
          ),
        )

 				jsonInput := []byte(`["vm-1234"]`)
 				err := json.Unmarshal(jsonInput, &extInput)
 				Expect(err).NotTo(HaveOccurred())

 				server.AppendHandlers(
 					helpers.MakeWorkflowHandlers(
 						"Deprovision",
 						cpiConfig.RequestID,
 						"57fb9fb03fcc55c807add41c",
 					)...,
 				)
 				server.AppendHandlers(
 					ghttp.CombineHandlers(
 						ghttp.VerifyRequest("DELETE", "/api/2.0/nodes/57fb9fb03fcc55c807add41c/tags/"+models.Unavailable),
 					  ghttp.RespondWith(http.StatusNoContent, nil),
          ),
 				)

 				err = cpi.DeleteVM(cpiConfig, extInput)
 				Expect(err).NotTo(HaveOccurred())
 				Expect(len(server.ReceivedRequests())).To(Equal(9))
 			})
 		})

 		//Context("when there are attached disks to a VM", func() {
 		//	It("detaches a disk from the VM and deletes the VM", func() {
 		//		jsonInput := []byte(`["valid_vm_cid_2"]`)
 		//		err := json.Unmarshal(jsonInput, &extInput)
 		//		Expect(err).NotTo(HaveOccurred())
 		//		nodeID := "5665a65a0561790005b77b85"
 		//		container := rackhdapi.PersistentDiskSettingsContainer{
 		//			PersistentDisk: rackhdapi.PersistentDiskSettings{
 		//				DiskCID:    fmt.Sprintf("%s-%s", nodeID, cpiConfig.RequestID),
 		//				Location:   fmt.Sprintf("/dev/%s", rackhdapi.PersistentDiskLocation),
 		//				IsAttached: false,
 		//			},
 		//		}
 		//		expectedPersistentDiskSettings, err := json.Marshal(container)
 		//		Expect(err).ToNot(HaveOccurred())
    //
 		//		server.AppendHandlers(
 		//			ghttp.CombineHandlers(
 		//				ghttp.VerifyRequest("PATCH", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
 		//				ghttp.VerifyJSON(string(expectedPersistentDiskSettings)),
 		//			),
 		//		)
 		//		server.AppendHandlers(
 		//			helpers.MakeWorkflowHandlers(
 		//				"Deprovision",
 		//				cpiConfig.RequestID,
 		//				nodeID,
 		//			)...,
 		//		)
 		//		err = cpi.DeleteVM(cpiConfig, extInput)
 		//		Expect(err).NotTo(HaveOccurred())
 		//		Expect(len(server.ReceivedRequests())).To(Equal(9))
 		//	})
 		//})
 	})
})
