package cpi_test

import (
	"github.com/rackhd/rackhd-cpi/cpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ImplementsMethod", func() {
	It("returns true if the CPI currently implements the method", func() {
		Expect(cpi.ImplementsMethod("create_vm")).To(BeTrue())
		Expect(cpi.ImplementsMethod("delete_vm")).To(BeTrue())
		Expect(cpi.ImplementsMethod("has_vm")).To(BeTrue())
		Expect(cpi.ImplementsMethod("create_stemcell")).To(BeTrue())
		Expect(cpi.ImplementsMethod("delete_stemcell")).To(BeTrue())
		Expect(cpi.ImplementsMethod("set_vm_metadata")).To(BeTrue())
		Expect(cpi.ImplementsMethod("delete_disk")).To(BeTrue())
		Expect(cpi.ImplementsMethod("attach_disk")).To(BeTrue())
		Expect(cpi.ImplementsMethod("detach_disk")).To(BeTrue())
		Expect(cpi.ImplementsMethod("has_disk")).To(BeTrue())
		Expect(cpi.ImplementsMethod("get_disks")).To(BeTrue())
		Expect(cpi.ImplementsMethod("create_disk")).To(BeTrue())
	})

	It("returns false if the CPI currently does not implement the method", func() {
		Expect(cpi.ImplementsMethod("reboot_vm")).To(BeFalse())
		Expect(cpi.ImplementsMethod("snapshot_disk")).To(BeFalse())
		Expect(cpi.ImplementsMethod("delete_snapshot")).To(BeFalse())
		Expect(cpi.ImplementsMethod("current_vm_id")).To(BeFalse())
		Expect(cpi.ImplementsMethod("configure_networks")).To(BeFalse())
	})

	It("returns an error if the method is invalid", func() {
		implemented, err := cpi.ImplementsMethod("some_fake_method")
		Expect(implemented).To(BeFalse())
		Expect(err).To(HaveOccurred())
	})
})
