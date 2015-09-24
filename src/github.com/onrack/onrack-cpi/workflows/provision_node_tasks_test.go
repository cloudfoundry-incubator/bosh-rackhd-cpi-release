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

var _ = Describe("CreateVmTasks", func() {
	Describe("ProvisionNodeTemplate", func() {
		It("has the expected stucture", func() {
			vendoredTask := ProvisionNodeTask{}
			err := json.Unmarshal(provisionNodeTemplate, &vendoredTask)
			Expect(err).ToNot(HaveOccurred())

			provisionNodeTaskFile, err := os.Open("../templates/provision_node_task.json")
			Expect(err).ToNot(HaveOccurred())
			defer provisionNodeTaskFile.Close()

			b, err := ioutil.ReadAll(provisionNodeTaskFile)
			Expect(err).ToNot(HaveOccurred())

			expectedTask := ProvisionNodeTask{}
			err = json.Unmarshal(b, &expectedTask)
			Expect(err).ToNot(HaveOccurred())

			Expect(vendoredTask).To(Equal(expectedTask))
		})

		It("marshalls into the expected JSON document", func() {
			vendoredTask := ProvisionNodeTask{}
			err := json.Unmarshal(provisionNodeTemplate, &vendoredTask)
			Expect(err).ToNot(HaveOccurred())

			vendoredTaskJSON, err := json.Marshal(vendoredTask)
			Expect(err).ToNot(HaveOccurred())

			provisionNodeTaskFile, err := os.Open("../templates/provision_node_task.json")
			Expect(err).ToNot(HaveOccurred())
			defer provisionNodeTaskFile.Close()

			expectedTaskJSON, err := ioutil.ReadAll(provisionNodeTaskFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(vendoredTaskJSON).To(MatchJSON(expectedTaskJSON))
		})
	})
})
