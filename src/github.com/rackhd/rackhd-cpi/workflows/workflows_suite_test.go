package workflows_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWorkflows(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workflows Suite")
}
