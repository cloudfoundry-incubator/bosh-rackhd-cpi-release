package cli_test

import (
	"github.com/onrack/onrack-cpi/cli"
	"github.com/onrack/onrack-cpi/cpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cli", func() {
	Describe("ParseCommand", func() {
		Context("With invalid method", func() {
			It("Returns invalid method", func() {
				testInput := []byte("invalid_method some-invalid-options")
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid Method: invalid_method"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(cpi.ExternalInput{}))
			})
		})
		Context("For unsupported CPI methods", func() {
			It("Dispatches reboot_vm", func() {
				testInput := []byte("reboot_vm some-broken-vm-cid")
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method reboot_vm is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(cpi.ExternalInput{}))
			})
			It("Dispatches reboot_vm", func() {
				testInput := []byte("set_vm_metadata some-broken-vm-cid")
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method set_vm_metadata is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(cpi.ExternalInput{}))
			})
			It("Dispatches configure_networks", func() {
				testInput := []byte("configure_networks some-broken-vm-cid")
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method configure_networks is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(cpi.ExternalInput{}))
			})
			It("Dispatches create_disk", func() {
				testInput := []byte("create_disk some-broken-vm-cid")
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method create_disk is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(cpi.ExternalInput{}))
			})
			It("Dispatches delete_disk", func() {
				testInput := []byte("delete_disk some-broken-vm-cid")
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method delete_disk is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(cpi.ExternalInput{}))
			})
			It("Dispatches attach_disk", func() {
				testInput := []byte("attach_disk some-broken-vm-cid")
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method attach_disk is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(cpi.ExternalInput{}))
			})
		})
		Context("For supported CPI methods", func() {
			It("Dispatches create_stemcell", func() {
				testInput := []byte(`create_stemcell ["some-awesome-stemcell-options"]`)
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.CREATE_STEMCELL))
				Expect(commandInput).To(Equal(cpi.ExternalInput{"some-awesome-stemcell-options"}))
			})
			It("Dispatches delete_stemcell", func() {
				testInput := []byte(`delete_stemcell ["some-useless-stemcell-cid"]`)
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.DELETE_STEMCELL))
				Expect(commandInput).To(Equal(cpi.ExternalInput{"some-useless-stemcell-cid"}))
			})
			It("Dispatches create_vm", func() {
				testInput := []byte(`create_vm ["some-awesome-vm-options"]`)
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.CREATE_VM))
				Expect(commandInput).To(Equal(cpi.ExternalInput{"some-awesome-vm-options"}))
			})
			It("Dispatches delete_vm", func() {
				testInput := []byte(`delete_vm ["some-unused-vm-cid"]`)
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.DELETE_VM))
				Expect(commandInput).To(Equal(cpi.ExternalInput{"some-unused-vm-cid"}))
			})
			It("Dispatches has_vm", func() {
				testInput := []byte(`has_vm ["some-interesting-vm-cid"]`)
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.HAS_VM))
				Expect(commandInput).To(Equal(cpi.ExternalInput{"some-interesting-vm-cid"}))
			})
		})
	})
})
