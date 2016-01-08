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
