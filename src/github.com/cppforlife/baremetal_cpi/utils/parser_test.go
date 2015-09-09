package utils_test

import (
	utils "github.com/cppforlife/baremetal_cpi/utils"
	"net/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"strings"
)

var _ = Describe("Parser", func() {


	Context("Test Parsing Response To Json", func() {
		It("Should convert the response to json", func() {
			body := ioutil.NopCloser(strings.NewReader(`{"name": "Test", "id": 123}`))
			resp := http.Response{"200", 200, "HTTP/1.0", 1, 0, nil, body, -1, nil, true, nil, nil, nil}

			var example ExampleJson
			err := utils.ReadResponseToJson(&resp, &example)
			Expect(err).ToNot(HaveOccurred())
			Expect(example.Name).To(Equal("Test"))
			Expect(example.Id).To(Equal(123))
		})

	})
})

type ExampleJson struct {
	Name string `json:"name"`
	Id int `json:"id"`
}