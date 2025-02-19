package httpclient

import (
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7"
	basehttpclient "github.com/gojek/heimdall/v7/httpclient"
)

type Client struct {
	client *basehttpclient.Client
}

var (
	initalTimeout                 = 2 * time.Millisecond // Inital timeout
	maxTimeout                    = 9 * time.Millisecond // Max time out
	exponentFactor        float64 = 2                    // Multiplier
	maximumJitterInterval         = 2 * time.Millisecond // Max jitter interval. It must be more than 1*time.Millisecond
)

func New(timeout time.Duration) *Client {
	backoff := heimdall.NewExponentialBackoff(
		initalTimeout,
		maxTimeout,
		exponentFactor,
		maximumJitterInterval)

	retrier := heimdall.NewRetrier(backoff)

	heimdallClient := basehttpclient.NewClient(
		basehttpclient.WithHTTPTimeout(timeout),
		basehttpclient.WithRetrier(retrier),
		basehttpclient.WithRetryCount(5),
	)

	client := &Client{
		client: heimdallClient,
	}

	return client
}

func NewWithClient(timeout time.Duration, client *http.Client) *Client {
	backoff := heimdall.NewExponentialBackoff(
		initalTimeout,
		maxTimeout,
		exponentFactor,
		maximumJitterInterval)

	retrier := heimdall.NewRetrier(backoff)

	heimdallClient := basehttpclient.NewClient(
		basehttpclient.WithHTTPClient(client),
		basehttpclient.WithHTTPTimeout(timeout),
		basehttpclient.WithRetrier(retrier),
		basehttpclient.WithRetryCount(5),
	)

	return &Client{
		client: heimdallClient,
	}
}

func (c *Client) Do(request *http.Request) (*http.Response, error) {
	return c.client.Do(request)
}
