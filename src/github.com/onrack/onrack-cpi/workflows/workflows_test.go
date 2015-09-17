package workflows_test

import (
	"encoding/json"
	"fmt"
	"github.com/onrack/onrack-cpi/workflows"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"

	"io/ioutil"
	"net/http"
	"github.com/nu7hatch/gouuid"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workflows", func() {
	Describe("PublishCreateVMWorkflow", func (){
		It("returns", func(){
			apiServer := os.Getenv("ON_RACK_API_URI")
			cpiConfig := config.Cpi{ApiServer: apiServer}
			fakeUUID, _ := uuid.NewV4()
			fakeUUIDstr := fakeUUID.String()

			err := workflows.PublishCreateVMWorkflow(cpiConfig, fakeUUIDstr)
			Expect(err).ToNot(HaveOccurred())

			expectedWorkflow := workflows.GenerateCreateVMWorkflow(fakeUUIDstr)
			expectedReserveNodeTask := workflows.GenerateReserveNodeTask(fakeUUIDstr)
			delete(expectedReserveNodeTask.Options, "commands")
			expectedProvisionNodeTask := workflows.GenerateProvisionNodeTask(fakeUUIDstr)
			delete(expectedProvisionNodeTask.Options, "commands")

			workflowLibrary, err := onrackhttp.RetrieveWorkflows(cpiConfig)
			taskLibrary, err := onrackhttp.RetrieveTasks(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			for i := range taskLibrary {
				delete(taskLibrary[i].Options, "commands")
			}
			Expect(taskLibrary).To(ContainElement(expectedProvisionNodeTask))
			Expect(taskLibrary).To(ContainElement(expectedReserveNodeTask))

			Expect(workflowLibrary).To(ContainElement(expectedWorkflow))
		})
	})
})
