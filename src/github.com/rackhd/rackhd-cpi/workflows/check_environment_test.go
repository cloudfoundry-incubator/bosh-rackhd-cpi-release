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
		apiServer := fmt.Sprintf("%s:%s", os.Getenv("RACKHD_API_HOST"), os.Getenv("RACKHD_API_PORT"))
		Expect(apiServer).ToNot(BeEmpty())

		c := config.Cpi{ApiServer: apiServer}
		err := workflows.BootstrappingTasksExist(c)
		Expect(err).ToNot(HaveOccurred())
	})
})
