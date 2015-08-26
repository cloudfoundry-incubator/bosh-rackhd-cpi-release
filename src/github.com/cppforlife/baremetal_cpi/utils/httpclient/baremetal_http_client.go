package httpclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
	"io"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var DefaultClient = http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

type HTTPClient interface {
	Put(endpoint string, content io.ReadCloser, contentLength int64) (*http.Response, error)
	Post(endpoint string, reader io.Reader) (*http.Response, error)
	Get(endpoint string) (*http.Response, error)
	Delete(endpoint string) (*http.Response, error)
}

type httpClient struct {
	client http.Client
	logger boshlog.Logger
	logTag string
}

func NewHTTPClient(client http.Client, logger boshlog.Logger) HTTPClient {
	return httpClient{
		client: client,
		logger: logger,
		logTag: "httpClient",
	}
}

func (c httpClient) Put(endpoint string, content io.ReadCloser, contentLength int64) (*http.Response, error) {
	c.logger.Debug(c.logTag, "Sending PUT request, endpoint %s", endpoint)

	request, err := http.NewRequest("PUT", endpoint, content)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating PUT request")
	}

	request.ContentLength = contentLength
	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing PUT request")
	}
	return response, nil
}

func (c httpClient) Post(endpoint string, reader io.Reader) (*http.Response, error) {
	c.logger.Debug(c.logTag, "Sending POST request with body, endpoint %s", endpoint)

	request, err := http.NewRequest("POST", endpoint, reader)
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, bosherr.WrapError(err, "Creating POST request")
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing POST request")
	}
	return response, nil
}

func (c httpClient) Get(endpoint string) (*http.Response, error) {
	response, err := http.Get(endpoint)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing GET request")
	}

	return response, nil
}

func (c httpClient) Delete(endpoint string) (*http.Response, error) {
	c.logger.Debug(c.logTag, "Sending DELETE request with endpoint %s", endpoint)

	request, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating DELETE request")
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing DELETE request")
	}
	return response, nil
}

