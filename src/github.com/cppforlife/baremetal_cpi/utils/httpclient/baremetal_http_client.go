package utils

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
		return nil, bosherr.WrapError(err, "Creating POST request")
	}

	request.ContentLength = contentLength
	response, err := c.client.Do(request)
	if err != nil {
		return nil, bosherr.WrapError(err, "Performing POST request")
	}
	return response, nil
}

