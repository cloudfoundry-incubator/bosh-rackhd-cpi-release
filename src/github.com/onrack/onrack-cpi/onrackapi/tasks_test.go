package onrackapi_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tasks", func() {
	Describe("Publishing tasks", func() {
		It("adds task to library, retrieves updated list of tasks from task library", func() {
			apiServer := os.Getenv("ON_RACK_API_URI")
			Expect(apiServer).ToNot(BeEmpty())

			uuidObj, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uuid := uuidObj.String()
			cpiConfig := config.Cpi{ApiServer: apiServer}

			fakeTaskStub := onrackapi.TaskStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackapi.DefaultUnusedName,
			}

			fakeTask := struct {
				*onrackapi.TaskStub
				*onrackapi.OptionContainer
			}{
				TaskStub:        &fakeTaskStub,
				OptionContainer: &onrackapi.OptionContainer{},
			}

			fakeTaskBytes, err := json.Marshal(fakeTask)
			Expect(err).ToNot(HaveOccurred())

			err = onrackapi.PublishTask(cpiConfig, fakeTaskBytes)
			Expect(err).ToNot(HaveOccurred())

			taskLibraryBytes, err := onrackapi.RetrieveTasks(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			taskLibrary := []onrackapi.TaskStub{}
			err = json.Unmarshal(taskLibraryBytes, &taskLibrary)
			Expect(err).ToNot(HaveOccurred())

			Expect(taskLibrary).To(ContainElement(fakeTaskStub))
		})
	})
})
