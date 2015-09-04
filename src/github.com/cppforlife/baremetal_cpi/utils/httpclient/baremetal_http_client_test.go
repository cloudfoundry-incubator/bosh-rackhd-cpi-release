package httpclient_test

import (
	"io/ioutil"
	"net"
	"net/http"

	. "github.com/cppforlife/baremetal_cpi/utils/httpclient"
	. "github.com/cloudfoundry/bosh-utils/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-utils/internal/github.com/onsi/gomega"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"strings"
)

var _ = Describe("HttpClient", func() {
	var (
		httpClient HTTPClient
		fakeServer *fakeServer
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		httpClient = NewHTTPClient(DefaultClient, logger)
		fakeServer = newFakeServer("localhost:6305")

		readyCh := make(chan error)
		go fakeServer.Start(readyCh)
		err := <-readyCh
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		fakeServer.Stop()
	})

	Describe("Put", func() {
		It("makes a put request with given content and content's length", func() {
			fakeServer.SetResponseBody("fake-put-response")
			fakeServer.SetResponseStatus(200)

			body := ioutil.NopCloser(strings.NewReader("content"))
			response, err := httpClient.Put("http://localhost:6305/fake-path", body, int64(7))
			Expect(err).ToNot(HaveOccurred())

			defer response.Body.Close()
			responseBody, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())

			Expect(responseBody).To(Equal([]byte("fake-put-response")))
			Expect(response.StatusCode).To(Equal(200))

			Expect(fakeServer.ReceivedRequests).To(HaveLen(1))
			Expect(fakeServer.ReceivedRequests).To(ContainElement(
				receivedRequest{
					Body:   []byte("content"),
					Method: "PUT",
				},
			))
		})
	})

	Describe("Get", func() {
		Context("make a GET request to a specific address", func() {
			It("should return the content at that request url", func() {
				fakeServer.SetResponseBody("fake-put-response")
				fakeServer.SetResponseStatus(200)

				response, err := httpClient.Get("http://localhost:6305/fake-path")
				Ω(err).ToNot(HaveOccurred())

				defer response.Body.Close()
				responseBody, err := ioutil.ReadAll(response.Body)
				Ω(responseBody).To(Equal([]byte("fake-put-response")))
				Ω(response.StatusCode).To(Equal(200))
			})
		})

	})

	Describe("Post", func() {
		Context("make a POST request with json body", func() {
			It("should have appropriate json body when the server receive it", func() {
				fakeServer.SetResponseBody("fake-put-response")
				fakeServer.SetResponseStatus(200)

				reader := strings.NewReader(`{"name":"test", "id":1}`)
				response, err := httpClient.Post("http://localhost:6305/fake-path", reader)
				Ω(err).ToNot(HaveOccurred())
				defer response.Body.Close()
				responseBody, err := ioutil.ReadAll(response.Body)
				Ω(responseBody).To(Equal([]byte("fake-put-response")))
				Ω(response.StatusCode).To(Equal(200))

				Ω(fakeServer.ReceivedRequests).To(HaveLen(1))
				Ω(fakeServer.ReceivedRequests).To(ContainElement(
					receivedRequest{
						Body:   []byte(`{"name":"test", "id":1}`),
						Method: "POST",
					},
				))
			})
		})

	})
})

type receivedRequestBody struct {
	Method    string
	Arguments []interface{}
	ReplyTo   string `json:"reply_to"`
}

type receivedRequest struct {
	Body   []byte
	Method string
}

type fakeServer struct {
	listener         net.Listener
	endpoint         string
	ReceivedRequests []receivedRequest
	responseBody     string
	responseStatus   int
}

func newFakeServer(endpoint string) *fakeServer {
	return &fakeServer{
		endpoint:         endpoint,
		responseStatus:   http.StatusOK,
		ReceivedRequests: []receivedRequest{},
	}
}

func (s *fakeServer) Start(readyErrCh chan error) {
	var err error
	s.listener, err = net.Listen("tcp", s.endpoint)
	if err != nil {
		readyErrCh <- err
		return
	}

	readyErrCh <- nil

	httpServer := http.Server{}
	httpServer.SetKeepAlivesEnabled(false)
	mux := http.NewServeMux()
	httpServer.Handler = mux

	mux.HandleFunc("/fake-path", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.responseStatus)

		requestBody, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		receivedRequest := receivedRequest{
			Body:   requestBody,
			Method: r.Method,
		}

		s.ReceivedRequests = append(s.ReceivedRequests, receivedRequest)
		w.Write([]byte(s.responseBody))
	})

	httpServer.Serve(s.listener)
}

func (s *fakeServer) Stop() {
	s.listener.Close()
}

func (s *fakeServer) SetResponseStatus(code int) {
	s.responseStatus = code
}

func (s *fakeServer) SetResponseBody(body string) {
	s.responseBody = body
}
