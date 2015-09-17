package onrackhttp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRequests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OnRack HTTP Suite")
}
