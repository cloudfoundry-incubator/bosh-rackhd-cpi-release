package workflows_test

import (
	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/helpers"
	"github.com/rackhd/rackhd-cpi/rackhdapi"
	"github.com/rackhd/rackhd-cpi/workflows"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckEnvironment", func() {
	It("Returns no error when run against a properly configured environment", func() {
		apiServer, err := helpers.GetRackHDHost()
		Expect(err).ToNot(HaveOccurred())
		c := config.Cpi{ApiServer: apiServer}

		requiredTasks := workflows.GetRequiredTasks()

		for taskName, templatePath := range requiredTasks {
			taskBytes, err := rackhdapi.GetTaskBytes(c, taskName)
			Expect(err).ToNot(HaveOccurred())

			expectedTaskBytes, err := helpers.ReadFile(templatePath)
			Expect(err).ToNot(HaveOccurred())

			expectedTaskBytes = []byte(`[` + string(expectedTaskBytes) + `]`)
			Expect(taskBytes).To(MatchJSON(expectedTaskBytes))
		}
	})
})
