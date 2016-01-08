package config_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rackhd/rackhd-cpi/config"
)

var _ = Describe("Creating a config", func() {
	It("checks that the API server URI is set", func() {
		jsonReader := strings.NewReader(`{}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError("ApiServer IP is not set"))
	})

	It("checks that the agent configuration has an mbus setting", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"some": "options"}}}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError(`Agent config invalid {map[some:options]  [] map[]}`))
	})

	It("checks that the agent configuration includes a blobstore section", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"mbus":"localhost"}}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError(`Agent config invalid {map[] localhost [] map[]}`))
	})

	It("checks that the agent configuration includes a disks section", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"mbus":"localhost", "blobstore":{"some": "options"}}}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError(`Agent config invalid {map[some:options] localhost [] map[]}`))
	})

	It("checks that the agent configuration includes a blobstore provider section", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"mbus":"localhost","blobstore":{"some": "options"}}}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError(`Agent config invalid {map[some:options] localhost [] map[]}`))
	})

	It("checks that the agent configuration includes a system disk section", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"mbus":"localhost","blobstore":{"provider": "local"}, "disks":{"some": "options"}}}`)
		_, err := config.New(jsonReader)
		Expect(err).To(MatchError(`Agent config invalid {map[provider:local] localhost [] map[some:options]}`))
	})

	It("sets a default for max_create_vm_attmpts if not provided", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"provider": "local", "some": "options"}, "mbus":"localhost", "disks": {"system": "/dev/sda"}}}`)
		c, err := config.New(jsonReader)
		Expect(err).ToNot(HaveOccurred())
		Expect(c.MaxCreateVMAttempt).To(Equal(config.DefaultMaxCreateVMAttempts()))
	})

	It("checks that the agent configuration includes a disks section with system disk", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080","agent": {"blobstore": {"provider": "local", "some": "options"},	"mbus":"localhost", "disks": {"system": "/dev/sda"}}}`)
		c, err := config.New(jsonReader)
		Expect(err).ToNot(HaveOccurred())
		Expect(c.Agent.Disks["system"]).To(Equal("/dev/sda"))
	})
})
