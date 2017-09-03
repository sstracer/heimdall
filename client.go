package heimdall

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const defaultRetryCount int = 0

// Client Is a generic client interface
type Client interface {
	Get(url string) (Response, error)
	Post(url string, body io.Reader) (Response, error)
	Put(url string, body io.Reader) (Response, error)
	Patch(url string, body io.Reader) (Response, error)
	Delete(url string) (Response, error)
}

type httpClient struct {
	client *http.Client

	retryCount int
	retrier    Retriable
}

// NewHTTPClient returns a new instance of HTTPClient
func NewHTTPClient(timeoutInSeconds int) Client {
	httpTimeout := time.Duration(timeoutInSeconds) * time.Second
	return &httpClient{
		client: &http.Client{
			Timeout: httpTimeout,
		},

		retryCount: defaultRetryCount,
		retrier:    NewNoRetrier(),
	}
}

// SetRetryCount sets the retry count for the httpClient
func (c *httpClient) SetRetryCount(count int) {
	c.retryCount = count
}

// SetRetrier sets the strategy for retrying
func (c *httpClient) SetRetrier(retrier Retriable) {
	c.retrier = retrier
}

// Get makes a HTTP GET request to provided URL
func (c *httpClient) Get(url string) (Response, error) {
	response := Response{}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return response, errors.Wrap(err, "GET - request creation failed")
	}

	return c.do(request)
}

// Post makes a HTTP POST request to provided URL and requestBody
func (c *httpClient) Post(url string, body io.Reader) (Response, error) {
	response := Response{}

	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return response, errors.Wrap(err, "POST - request creation failed")
	}

	return c.do(request)
}

// Put makes a HTTP PUT request to provided URL and requestBody
func (c *httpClient) Put(url string, body io.Reader) (Response, error) {
	response := Response{}

	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return response, errors.Wrap(err, "PUT - request creation failed")
	}

	return c.do(request)
}

// Patch makes a HTTP PATCH request to provided URL and requestBody
func (c *httpClient) Patch(url string, body io.Reader) (Response, error) {
	response := Response{}

	request, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		return response, errors.Wrap(err, "PATCH - request creation failed")
	}

	return c.do(request)
}

// Delete makes a HTTP DELETE request with provided URL
func (c *httpClient) Delete(url string) (Response, error) {
	response := Response{}

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return response, errors.Wrap(err, "DELETE - request creation failed")
	}

	return c.do(request)
}

func (c *httpClient) do(request *http.Request) (Response, error) {
	hr := Response{}

	request.Close = true

	for i := 0; i <= c.retryCount; i++ {

		response, err := c.client.Do(request)
		if err != nil {
			backoffTime := c.retrier.NextInterval(i)
			time.Sleep(backoffTime)
			continue
		}

		defer response.Body.Close()

		var responseBytes []byte
		if response.Body != nil {
			responseBytes, err = ioutil.ReadAll(response.Body)
			if err != nil {
				backoffTime := c.retrier.NextInterval(i)
				time.Sleep(backoffTime)
				continue
			}
		}

		hr.body = responseBytes
		hr.statusCode = response.StatusCode
	}

	return hr, nil
}