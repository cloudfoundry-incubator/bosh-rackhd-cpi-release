package bosh_test

import (
	"io/ioutil"
	"testing"

	log "github.com/Sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBosh(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bosh Suite")
}
