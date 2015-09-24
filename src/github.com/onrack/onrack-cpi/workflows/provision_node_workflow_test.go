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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProvisionNodeWorkflow", func() {
	Describe("ProvisionNodeWorkflow", func() {
		It("has the expected structure", func() {
			vendoredWorkflow := ProvisionNodeWorkflow{}
			err := json.Unmarshal(provisionNodeWorkflowTemplate, &vendoredWorkflow)
			Expect(err).ToNot(HaveOccurred())

			provisionNodeWorkflowFile, err := os.Open("../templates/provision_node_workflow.json")
			Expect(err).ToNot(HaveOccurred())
			defer provisionNodeWorkflowFile.Close()

			b, err := ioutil.ReadAll(provisionNodeWorkflowFile)
			Expect(err).ToNot(HaveOccurred())

			expectedWorkflow := ProvisionNodeWorkflow{}
			err = json.Unmarshal(b, &expectedWorkflow)
			Expect(err).ToNot(HaveOccurred())

			Expect(vendoredWorkflow).To(Equal(expectedWorkflow))
		})

		It("marshalls into the expected JSON document", func() {
			vendoredWorkflow := ProvisionNodeWorkflow{}
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
})
