package workflows_test

import (
	"os"

	"github.com/onrack/onrack-cpi/config"
	"github.com/onrack/onrack-cpi/workflows"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckEnvironment", func() {
	It("Returns no error when run against a properly configured environment", func() {
		apiServerIP := os.Getenv("ON_RACK_API_URI")
		Expect(apiServerIP).ToNot(BeEmpty())

		c := config.Cpi{ApiServer: apiServerIP}
		err := workflows.BootstrappingTasksExist(c)
		Expect(err).ToNot(HaveOccurred())
	})
})
