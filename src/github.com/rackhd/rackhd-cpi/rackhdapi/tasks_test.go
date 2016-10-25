package rackhdapi_test

import (
  "encoding/json"
  "fmt"

  "github.com/rackhd/rackhd-cpi/config"
  "github.com/rackhd/rackhd-cpi/helpers"
  "github.com/rackhd/rackhd-cpi/models"
  "github.com/rackhd/rackhd-cpi/rackhdapi"

  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
)

var _ = Describe("Tasks", func() {
  Describe("Publishing tasks", func() {
    It("adds task to library, retrieves updated list of tasks from task library", func() {
      uuid, err := helpers.GenerateUUID()
      Expect(err).ToNot(HaveOccurred())

      apiServer, err := helpers.GetRackHDHost()
      Expect(err).ToNot(HaveOccurred())
      cpiConfig := config.Cpi{ApiServer: apiServer}

      fakeTask := models.Task{
        Name:           fmt.Sprintf("Task.CF.Fake.%s", uuid),
        UnusedName:     models.DefaultUnusedName,
        ImplementsTask: "Task.Base.Linux.Commands",
        Options:        map[string]interface{}{},
        Properties:     models.TaskProperties{},
      }
      fakeTaskBytes, err := json.Marshal(fakeTask)
      Expect(err).ToNot(HaveOccurred())

      err = rackhdapi.PublishTask(cpiConfig, fakeTaskBytes)
      Expect(err).ToNot(HaveOccurred())

      publishedTask, err := rackhdapi.RetrieveTask(cpiConfig, fakeTask.Name)
      Expect(err).ToNot(HaveOccurred())

      publishedTaskBytes, err := json.Marshal(publishedTask)
      Expect(err).ToNot(HaveOccurred())
      Expect(publishedTaskBytes).To(MatchJSON(fakeTaskBytes))
    })
  })
})
