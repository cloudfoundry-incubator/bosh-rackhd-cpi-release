package cpi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rackhd/rackhd-cpi/bosh"
	"github.com/rackhd/rackhd-cpi/config"
	. "github.com/rackhd/rackhd-cpi/cpi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/rackhd/rackhd-cpi/helpers"
)

var _ = Describe("DeleteDisk", func() {

	var server *ghttp.Server
	var jsonReader *strings.Reader
	var cpiConfig config.Cpi

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverURL, err := url.Parse(server.URL())
		Expect(err).ToNot(HaveOccurred())
		jsonReader = strings.NewReader(fmt.Sprintf(`{"apiserver":"%s", "agent":{"blobstore": {"provider":"local","some": "options"}, "mbus":"localhost"}, "max_create_vm_attempts":1}`, serverURL.Host))
		cpiConfig, err = config.New(jsonReader)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("when given a disk cid for an existing disk", func() {
		It("deletes the disk", func() {
			jsonInput := []byte(`[
					"valid_disk_cid_1"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).NotTo(HaveOccurred())

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_disks_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
				ghttp.VerifyRequest("PATCH", "/api/common/nodes/55e79ea54e66816f6152fff9"),
			)

			err = DeleteDisk(cpiConfig, extInput)
			Expect(len(server.ReceivedRequests())).To(Equal(2))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when given a disk cid for a non-existent disk", func() {
		It("returns an error", func() {
			jsonInput := []byte(`[
					"invalid_disk_cid"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_disks_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			err = DeleteDisk(cpiConfig, extInput)
			Expect(err).To(MatchError("Disk: invalid_disk_cid not found\n"))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})

	Context("when given a disk cid for a attached disk", func() {
		It("returns an error", func() {
			jsonInput := []byte(`[
					"valid_disk_cid_2"
				]`)
			var extInput bosh.MethodArguments
			err := json.Unmarshal(jsonInput, &extInput)
			Expect(err).ToNot(HaveOccurred())

			expectedNodes := helpers.LoadNodes("../spec_assets/dummy_disks_response.json")
			expectedNodesData, err := json.Marshal(expectedNodes)
			Expect(err).ToNot(HaveOccurred())
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/common/nodes"),
					ghttp.RespondWith(http.StatusOK, expectedNodesData),
				),
			)

			err = DeleteDisk(cpiConfig, extInput)
			Expect(err).To(MatchError("Disk: valid_disk_cid_2 is attached\n"))
			Expect(len(server.ReceivedRequests())).To(Equal(1))
		})
	})
})
