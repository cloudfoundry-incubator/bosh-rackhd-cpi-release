package rackhdapi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRequests(t *testing.T) {
	// where did my logs go
	// disable logging
	// log.SetOutput(ioutil.Discard)

	RegisterFailHandler(Fail)
	RunSpecs(t, "RackHD API Suite")
}
