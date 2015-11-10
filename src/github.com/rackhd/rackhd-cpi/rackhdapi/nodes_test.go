package rackhdapi_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rackhd/rackhd-cpi/config"
	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nodes", func() {
	Describe("Getting nodes", func() {
		It("return expected nodes fields", func() {
			apiServerIP := os.Getenv("RACKHD_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}

			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())

			resp, err := http.Get(fmt.Sprintf("http://%s:8080/api/common/nodes", c.ApiServer))
			Expect(err).ToNot(HaveOccurred())

			b, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var rawNodes []rackhdapi.Node
			err = json.Unmarshal(b, &rawNodes)
			Expect(err).ToNot(HaveOccurred())

			Expect(nodes).To(Equal(rawNodes))
		})
	})

	Describe("Getting catalog", func() {
		It("return ", func() {
			apiServerIP := os.Getenv("RACKHD_API_URI")
			Expect(apiServerIP).ToNot(BeEmpty())
			c := config.Cpi{ApiServer: apiServerIP}

			nodes, err := rackhdapi.GetNodes(c)
			Expect(err).ToNot(HaveOccurred())

			Expect(nodes).ToNot(BeEmpty())
			testNode := nodes[0]

			catalog, err := rackhdapi.GetNodeCatalog(c, testNode.ID)
			Expect(err).ToNot(HaveOccurred())

			resp, err := http.Get(fmt.Sprintf("http://%s:8080/api/common/nodes/%s/catalogs/ohai", c.ApiServer, testNode.ID))
			Expect(err).ToNot(HaveOccurred())

			b, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			var rawCatalog rackhdapi.NodeCatalog
			err = json.Unmarshal(b, &rawCatalog)
			Expect(err).ToNot(HaveOccurred())

			Expect(catalog).To(Equal(rawCatalog))
		})
	})
})
