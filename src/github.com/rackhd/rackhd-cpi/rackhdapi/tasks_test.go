package rackhdapi_test

import (
	"encoding/json"
	"fmt"

	"github.com/nu7hatch/gouuid"
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tasks", func() {
	Describe("Publishing tasks", func() {
		It("adds task to library, retrieves updated list of tasks from task library", func() {
			apiServer, err := helpers.GetRackHDHost()
			Expect(err).ToNot(HaveOccurred())

			uuidObj, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())
			uuid := uuidObj.String()
			cpiConfig := config.Cpi{ApiServer: apiServer}

			fakeTask := rackhdapi.Task{
				Name:           fmt.Sprintf("Task.CF.Fake.%s", uuid),
				UnusedName:     rackhdapi.DefaultUnusedName,
				ImplementsTask: "Task.Base.Node.Update",
			}

			fakeTaskBytes, err := json.Marshal(fakeTask)
			Expect(err).ToNot(HaveOccurred())

			err = rackhdapi.PublishTask(cpiConfig, fakeTaskBytes)
			Expect(err).ToNot(HaveOccurred())

			taskLibraryBytes, err := rackhdapi.RetrieveTasks(cpiConfig)
			Expect(err).ToNot(HaveOccurred())

			taskLibrary := []rackhdapi.Task{}
			err = json.Unmarshal(taskLibraryBytes, &taskLibrary)
			Expect(err).ToNot(HaveOccurred())

			Expect(taskLibrary).To(ContainElement(fakeTask))
		})
	})
})
