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
	"io/ioutil"
	"os"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
)

var _ = Describe("ProvisionNodeWorkflow", func() {
	Describe("ProvisionNodeWorkflow", func() {
		It("has the expected structure", func() {
			vendoredWorkflow := provisionNodeWorkflow{}
			err := json.Unmarshal(provisionNodeWorkflowTemplate, &vendoredWorkflow)
			Expect(err).ToNot(HaveOccurred())

			provisionNodeWorkflowFile, err := os.Open("../templates/provision_node_workflow.json")
			Expect(err).ToNot(HaveOccurred())
			defer provisionNodeWorkflowFile.Close()

			b, err := ioutil.ReadAll(provisionNodeWorkflowFile)
			Expect(err).ToNot(HaveOccurred())

			expectedWorkflow := provisionNodeWorkflow{}
			err = json.Unmarshal(b, &expectedWorkflow)
			Expect(err).ToNot(HaveOccurred())

			Expect(vendoredWorkflow).To(Equal(expectedWorkflow))
		})

		It("marshalls into the expected JSON document", func() {
			vendoredWorkflow := provisionNodeWorkflow{}
			err := json.Unmarshal(provisionNodeWorkflowTemplate, &vendoredWorkflow)
			Expect(err).ToNot(HaveOccurred())

			vendoredWorkflowJSON, err := json.Marshal(vendoredWorkflow)
			Expect(err).ToNot(HaveOccurred())

			provisionNodeWorkflowFile, err := os.Open("../templates/provision_node_workflow.json")
			Expect(err).ToNot(HaveOccurred())
			defer provisionNodeWorkflowFile.Close()

			expectedWorkflowJSON, err := ioutil.ReadAll(provisionNodeWorkflowFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(vendoredWorkflowJSON).To(MatchJSON(expectedWorkflowJSON))
		})
	})

	Describe("generating the set of provision workflow tasks and workflow", func() {
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

			s := setNodeIDTask{}
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

	Describe("publishing generated provision node tasks and workflow", func() {
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
})
