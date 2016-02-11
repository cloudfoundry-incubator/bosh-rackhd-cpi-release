package config_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
)

var _ = Describe("Creating a config", func() {
	var request bosh.CpiRequest

	BeforeEach(func() {
		request = bosh.CpiRequest{}
	})

	It("checks that the API server URI is set", func() {
		jsonReader := strings.NewReader(`{}`)
		_, err := config.New(jsonReader, request)
		Expect(err).To(MatchError("ApiServer IP is not set"))
	})

	It("checks that the agent configuration has an mbus setting", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"some": "options"}}}`)
		_, err := config.New(jsonReader, request)
		Expect(err).To(MatchError(`Agent config invalid {map[some:options]  []}`))
	})

	It("checks that the agent configuration includes a blobstore section", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"mbus":"localhost"}}`)
		_, err := config.New(jsonReader, request)
		Expect(err).To(MatchError(`Agent config invalid {map[] localhost []}`))
	})

	It("checks that the agent configuration includes a blobstore provider section", func() {
		jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"mbus":"localhost","blobstore":{"some": "options"}}}`)
		_, err := config.New(jsonReader, request)
		Expect(err).To(MatchError(`Agent config invalid {map[some:options] localhost []}`))
	})

	Context("with a request to create a VM", func() {
		It("sets a default for max_create_vm_attmpts if not provided", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"provider": "local", "some": "options"}, "mbus":"localhost"}}`)
			request.Method = bosh.CREATE_VM
			c, err := config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.MaxReserveNodeAttempts).To(Equal(config.DefaultMaxReserveNodeAttempts()))
		})
	})

	Context("with a request to create a VM", func() {
		It("sets a default for max_create_vm_attmpts if not provided", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"provider": "local", "some": "options"}, "mbus":"localhost"}}`)
			request.Method = bosh.CREATE_DISK
			c, err := config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.MaxReserveNodeAttempts).To(Equal(config.DefaultMaxReserveNodeAttempts()))
		})
	})

	Context("without a request to create a VM or disk", func() {
		It("sets max_create_vm_attmpts to zero if not provided", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"provider": "local", "some": "options"}, "mbus":"localhost"}}`)
			c, err := config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.MaxReserveNodeAttempts).To(Equal(0))
		})
	})

	Context("when uuid is not set", func() {
		It("generates a new one", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"provider": "local", "some": "options"}, "mbus":"localhost"}}`)
			c, err := config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RequestID).ToNot(Equal(""))
		})
	})

	Context("when uuid is set", func() {
		It("uses the specified uuid", func() {
			jsonReader := strings.NewReader(`{"apiserver":"localhost:8080", "agent":{"blobstore": {"provider": "local", "some": "options"}, "mbus":"localhost"}, "request_id": "9999"}`)
			c, err := config.New(jsonReader, request)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RequestID).To(Equal("9999"))
		})
	})
})
