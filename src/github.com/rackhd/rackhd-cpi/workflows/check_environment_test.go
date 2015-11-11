package workflows_test

import (
	"fmt"
	"os"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/workflows"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckEnvironment", func() {
	It("Returns no error when run against a properly configured environment", func() {
		apiServerIP := fmt.Sprintf("%s:8080", os.Getenv("RACKHD_API_URI"))
		Expect(apiServerIP).ToNot(BeEmpty())

		c := config.Cpi{ApiServer: apiServerIP}
		err := workflows.BootstrappingTasksExist(c)
		Expect(err).ToNot(HaveOccurred())
	})
})
