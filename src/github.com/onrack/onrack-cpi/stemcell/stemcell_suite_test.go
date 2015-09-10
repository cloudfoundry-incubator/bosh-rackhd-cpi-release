package stemcell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"log"
	"testing"
)

func TestStemcell(t *testing.T) {
	// where did my logs go
	// disable logging
        log.SetOutput(ioutil.Discard)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Stemcell Suite")
}
