package config_test

import (
	"strings"

	"github.com/onrack/onrack-cpi/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creating a config", func() {
	It("checks that the API server URI is set", func() {
		jsonReader := strings.NewReader(`{}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError("ApiServer IP is not set"))
	})

	It("checks that the agent configuration has an mbus setting", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost", "agent":{"blobstore": {"some": "options"}}}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError(`Agent config invalid {map[some:options]  []}`))
	})

	It("checks that the agent configuration includes a blobstore section", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost", "agent":{"mbus":"localhost"}}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError(`Agent config invalid {map[] localhost []}`))
	})

	It("sets a default for max_create_vm_attmpts if not provided", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost", "agent":{"blobstore": {"some": "options"}, "mbus":"localhost"}}`)
		c, err := config.New(jsonReader)
		Expect(err).ToNot(HaveOccurred())
		Expect(c.MaxCreateVMAttempt).To(Equal(config.DefaultMaxCreateVMAttempts()))
	})
})
