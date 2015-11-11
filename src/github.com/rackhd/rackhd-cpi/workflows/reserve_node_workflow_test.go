package workflows

/*
	Rather than creating a separate workflows_test package, this suite is part of
	the workflows package itself in order to provide regression tests for the ReserveNodeTasks
	vm task templates without exporting these variables for testing. Please be careful
	as this suite will have access to all unexported functions and variables in the workflows
	package. You have been warned

	- The ghost in the air ducts
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rackhd/rackhd-cpi/config"
)

var _ = Describe("ReserveNodeWorkflow", func() {
	Describe("ReserveNodeWorkflow", func() {
		It("have the expected structure", func() {
			vendoredWorkflow := reserveNodeWorkflow{}

			err := json.Unmarshal(reserveNodeWorkflowTemplate, &vendoredWorkflow)
			Expect(err).ToNot(HaveOccurred())

			reserveNodeWorkflowFile, err := os.Open("../templates/reserve_node_workflow.json")
			Expect(err).ToNot(HaveOccurred())
			defer reserveNodeWorkflowFile.Close()

			b, err := ioutil.ReadAll(reserveNodeWorkflowFile)
			Expect(err).ToNot(HaveOccurred())

			expectedWorkflow := reserveNodeWorkflow{}
			err = json.Unmarshal(b, &expectedWorkflow)
			Expect(err).ToNot(HaveOccurred())

			Expect(vendoredWorkflow).To(Equal(expectedWorkflow))
		})

		It("marshalls into the expected JSON document", func() {
			vendoredWorkflow := reserveNodeWorkflow{}
			err := json.Unmarshal(reserveNodeWorkflowTemplate, &vendoredWorkflow)
			Expect(err).ToNot(HaveOccurred())

			vendoredWorkflowJSON, err := json.Marshal(vendoredWorkflow)
			Expect(err).ToNot(HaveOccurred())

			reserveNodeWorkflowFile, err := os.Open("../templates/reserve_node_workflow.json")
			Expect(err).ToNot(HaveOccurred())
			defer reserveNodeWorkflowFile.Close()

			expectedWorkflowJSON, err := ioutil.ReadAll(reserveNodeWorkflowFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(vendoredWorkflowJSON).To(MatchJSON(expectedWorkflowJSON))
		})
	})

	Describe("generating the set of provision workflow tasks and workflow", func() {
		It("generates the required tasks and workflow with unique names", func() {
			u, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uID := u.String()

			tasksBytes, wBytes, err := generateReserveNodeWorkflow(uID)
			Expect(err).ToNot(HaveOccurred())

			r := reserveNodeTask{}
			err = json.Unmarshal(tasksBytes[0], &r)
			Expect(err).ToNot(HaveOccurred())
			Expect(r.Name).To(ContainSubstring(uID))

			w := reserveNodeWorkflow{}
			err = json.Unmarshal(wBytes, &w)
			Expect(err).ToNot(HaveOccurred())
			Expect(w.Tasks).To(HaveLen(4))
			Expect(w.Name).To(ContainSubstring(uID))

			Expect(w.Tasks[0].TaskName).To(Equal("Task.Obm.Node.PxeBoot"))
			Expect(w.Tasks[1].TaskName).To(Equal("Task.Obm.Node.Reboot"))
			Expect(w.Tasks[2].TaskName).To(Equal("Task.Linux.Bootstrap.Ubuntu"))
			Expect(w.Tasks[3].TaskName).To(Equal(r.Name))
		})
	})

	Describe("publishing generated reserve node workflow and tasks", func() {
		It("publishes the tasks and workflow", func() {
			u, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uID := u.String()

			apiServerIP := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}

			workflowName, err := PublishReserveNodeWorkflow(c, uID)
			Expect(err).ToNot(HaveOccurred())
			Expect(workflowName).To(ContainSubstring(uID))
		})
	})
})
