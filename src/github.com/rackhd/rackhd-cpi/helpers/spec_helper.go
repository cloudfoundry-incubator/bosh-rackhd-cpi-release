package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/rackhd/rackhd-cpi/rackhdapi"

	. "github.com/onsi/gomega"
)

func LoadJSON(nodePath string) []byte {
	dummyResponseFile, err := os.Open(nodePath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyResponseFile.Close()

	dummyResponseBytes, err := ioutil.ReadAll(dummyResponseFile)
	Expect(err).ToNot(HaveOccurred())

	return dummyResponseBytes
}

func LoadNodes(nodePath string) []rackhdapi.Node {
	dummyResponseBytes := LoadJSON(nodePath)

	nodes := []rackhdapi.Node{}
	err := json.Unmarshal(dummyResponseBytes, &nodes)
	Expect(err).ToNot(HaveOccurred())

	return nodes
}

func LoadNodeCatalog(nodeCatalogPath string) rackhdapi.NodeCatalog {
	dummyCatalogfile, err := os.Open(nodeCatalogPath)
	Expect(err).ToNot(HaveOccurred())
	defer dummyCatalogfile.Close()

	b, err := ioutil.ReadAll(dummyCatalogfile)
	Expect(err).ToNot(HaveOccurred())

	nodeCatalog := rackhdapi.NodeCatalog{}

	err = json.Unmarshal(b, &nodeCatalog)
	Expect(err).ToNot(HaveOccurred())
	return nodeCatalog
}
