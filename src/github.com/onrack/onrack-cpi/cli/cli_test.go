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
				testInput := "invalid_method some-invalid-options"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Invalid Method: invalid_method"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(""))
			})
		})
		Context("For unsupported CPI methods", func() {
			It("Dispatches reboot_vm", func() {
				testInput := "reboot_vm some-broken-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method reboot_vm is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(""))
			})
			It("Dispatches reboot_vm", func() {
				testInput := "set_vm_metadata some-broken-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method set_vm_metadata is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(""))
			})
			It("Dispatches configure_networks", func() {
				testInput := "configure_networks some-broken-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method configure_networks is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(""))
			})
			It("Dispatches create_disk", func() {
				testInput := "create_disk some-broken-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method create_disk is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(""))
			})
			It("Dispatches delete_disk", func() {
				testInput := "delete_disk some-broken-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method delete_disk is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(""))
			})
			It("Dispatches attach_disk", func() {
				testInput := "attach_disk some-broken-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("Method attach_disk is not implemented"))
				Expect(command).To(Equal(""))
				Expect(commandInput).To(Equal(""))
			})
		})
		Context("For supported CPI methods", func() {
			It("Dispatches create_stemcell", func() {
				testInput := "create_stemcell some-awesome-stemcell-options"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.CREATE_STEMCELL))
				Expect(commandInput).To(Equal("some-awesome-stemcell-options"))
			})
			It("Dispatches delete_stemcell", func() {
				testInput := "delete_stemcell some-useless-stemcell-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.DELETE_STEMCELL))
				Expect(commandInput).To(Equal("some-useless-stemcell-cid"))
			})
			It("Dispatches create_vm", func() {
				testInput := "create_vm some-awesome-vm-options"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.CREATE_VM))
				Expect(commandInput).To(Equal("some-awesome-vm-options"))
			})
			It("Dispatches delete_vm", func() {
				testInput := "delete_vm some-unused-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.DELETE_VM))
				Expect(commandInput).To(Equal("some-unused-vm-cid"))
			})
			It("Dispatches has_vm", func() {
				testInput := "has_vm some-interesting-vm-cid"
				command, commandInput, err := cli.ParseCommand(testInput)
				Expect(err).ToNot(HaveOccurred())
				Expect(command).To(Equal(cpi.HAS_VM))
				Expect(commandInput).To(Equal("some-interesting-vm-cid"))
			})
		})
	})
})
