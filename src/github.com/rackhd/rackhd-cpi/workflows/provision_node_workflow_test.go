package workflows

/*
  Rather than creating a separate workflows_test package, this suite is part of
  the workflows package itself in order to provide regression tests for the ProvisionNodeTasks
  vm task templates without exporting these variables for testing. Please be careful
  as this suite will have access to all unexported functions and variables in the workflows
  package. You have been warned

  - The ghost in the air ducts
*/

import (
  "encoding/json"
  "fmt"
  "net/http"

  "github.com/nu7hatch/gouuid"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
  "github.com/onsi/gomega/ghttp"
  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/helpers"
  "github.com/rackhd/rackhd-cpi/models"
)

var _ = Describe("ProvisionNodeWorkflow", func() {
  Describe("PublishProvisionNodeWorkflow", func() {
    It("publishes the tasks and workflow", func() {
      u, err := uuid.NewV4()
      Expect(err).ToNot(HaveOccurred())
      uID := u.String()

      apiServer, err := helpers.GetRackHDHost()
      Expect(err).ToNot(HaveOccurred())

      c := config.Cpi{ApiServer: apiServer, RequestID: uID}

      workflowName, err := PublishProvisionNodeWorkflow(c)
      Expect(err).ToNot(HaveOccurred())
      Expect(workflowName).To(ContainSubstring(uID))
    })
  })

  Describe("generateProvisionNodeWorkflow", func() {
    It("generates the required tasks and workflow with unique names", func() {
      u, err := uuid.NewV4()
      Expect(err).ToNot(HaveOccurred())
      uID := u.String()

      tasksBytes, wBytes, err := generateProvisionNodeWorkflow(uID)
      Expect(err).ToNot(HaveOccurred())

      p := provisionNodeTask{}
      err = json.Unmarshal(tasksBytes[0], &p)
      Expect(err).ToNot(HaveOccurred())
      Expect(p.Name).To(ContainSubstring(uID))

      s := models.Task{}
      err = json.Unmarshal(tasksBytes[1], &s)
      Expect(err).ToNot(HaveOccurred())
      Expect(s.Name).To(ContainSubstring(uID))

      w := provisionNodeWorkflow{}
      err = json.Unmarshal(wBytes, &w)
      Expect(err).ToNot(HaveOccurred())

      Expect(w.Name).To(ContainSubstring(uID))
      Expect(w.Tasks).To(HaveLen(4))
      Expect(w.Tasks[0].TaskName).To(Equal("Task.Linux.Bootstrap.Ubuntu"))
      Expect(w.Tasks[1].TaskName).To(Equal(p.Name))
      Expect(w.Tasks[2].TaskName).To(Equal(s.Name))
      Expect(w.Tasks[3].TaskName).To(Equal("Task.ProcShellReboot"))
    })
  })

  Describe("buildProvisionWorkflowOptions", func() {
    var server *ghttp.Server
    var cpiConfig config.Cpi

    BeforeEach(func() {
      server, _, cpiConfig, _ = helpers.SetUp("")
    })

    Context("when the node uses IPMI", func() {
      It("sets the OBM settings to IPMI", func() {
        expectedNode := helpers.LoadNode("../spec_assets/dummy_one_node_with_ipmi_response.json")
        expectedNodeData, err := json.Marshal(expectedNode)
        Expect(err).ToNot(HaveOccurred())

        nodeID := "5665a65a0561790005b77b85"
        server.AppendHandlers(
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
            ghttp.RespondWith(http.StatusOK, expectedNodeData),
          ),
        )

        envPath := models.RackHDEnvPath
        vmCID := "vmCID"
        stemcellCID := "stemcellCID"
        wipeDisk := "false"
        ipmiServiceName := models.OBMSettingIPMIServiceName
        expectedOptions := ProvisionNodeWorkflowOptions{
          AgentSettingsFile: &nodeID,
          AgentSettingsPath: &envPath,
          CID:               &vmCID,
          StemcellFile:      &stemcellCID,
          WipeDisk:          wipeDisk,
          OBMServiceName:    &ipmiServiceName,
        }

        options, err := buildProvisionWorkflowOptions(cpiConfig, nodeID, vmCID, stemcellCID, false)
        Expect(err).ToNot(HaveOccurred())
        Expect(options).To(Equal(expectedOptions))
      })
    })

    Context("when the node uses AMT", func() {
      It("sets the OMB settings to AMT", func() {
        expectedNode := helpers.LoadNode("../spec_assets/dummy_one_node_response.json")
        expectedNodeData, err := json.Marshal(expectedNode)
        Expect(err).ToNot(HaveOccurred())

        nodeID := "5665a65a0561790005b77b85"
        server.AppendHandlers(
          ghttp.CombineHandlers(
            ghttp.VerifyRequest("GET", fmt.Sprintf("/api/2.0/nodes/%s", nodeID)),
            ghttp.RespondWith(http.StatusOK, expectedNodeData),
          ),
        )

        envPath := models.RackHDEnvPath
        vmCID := "vmCID"
        stemcellCID := "stemcellCID"
        wipeDisk := "false"
        ipmiServiceName := models.OBMSettingAMTServiceName
        expectedOptions := ProvisionNodeWorkflowOptions{
          AgentSettingsFile: &nodeID,
          AgentSettingsPath: &envPath,
          CID:               &vmCID,
          StemcellFile:      &stemcellCID,
          WipeDisk:          wipeDisk,
          OBMServiceName:    &ipmiServiceName,
        }

        options, err := buildProvisionWorkflowOptions(cpiConfig, nodeID, vmCID, stemcellCID, false)
        Expect(err).ToNot(HaveOccurred())
        Expect(options).To(Equal(expectedOptions))
      })
    })
  })
})
