package onrackhttp_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nu7hatch/gouuid"
	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/onrackhttp"

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

			fakeTaskStub := onrackhttp.TaskStub{
				Name:       fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName: onrackhttp.DefaultUnusedName,
			}

			fakeTask := struct {
				*onrackhttp.TaskStub
				*onrackhttp.OptionContainer
			}{
				TaskStub:        &fakeTaskStub,
				OptionContainer: &onrackhttp.OptionContainer{},
			}

			fakeTaskBytes, err := json.Marshal(fakeTask)
			Expect(err).ToNot(HaveOccurred())

			err = onrackhttp.PublishTask(cpiConfig, fakeTaskBytes)
			Expect(err).ToNot(HaveOccurred())

			taskLibraryBytes, err := onrackhttp.RetrieveTasks(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			taskLibrary := []onrackhttp.TaskStub{}
			err = json.Unmarshal(taskLibraryBytes, &taskLibrary)
			Expect(err).ToNot(HaveOccurred())

			Expect(taskLibrary).To(ContainElement(fakeTaskStub))
		})
	})
})
