// Package plunk provides a Go client for the Plunk API.
//
// See https://docs.useplunk.com for API documentation.
package plunk

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://next-api.useplunk.com"

// Option configures a [Client].
type Option interface {
	apply(*Client)
}

// OptionFunc is an adapter to allow the use of ordinary functions as [Option]s.
type OptionFunc func(*Client)

func (f OptionFunc) apply(c *Client) { f(c) }

// WithBaseURL returns an [Option] that sets the base URL of the API. Trailing
// slashes are trimmed.
func WithBaseURL(url string) Option {
	return OptionFunc(func(c *Client) {
		c.BaseURL = strings.TrimRight(url, "/")
	})
}

// WithHTTPClient returns an [Option] that sets the HTTP client used for
// requests.
func WithHTTPClient(hc *http.Client) Option {
	return OptionFunc(func(c *Client) {
		c.HTTPClient = hc
	})
}

// Client is a Plunk API client. Use [New] to create one.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Retry      RetryConfig
}

// New creates a new Plunk API client with the given API key and options.
//
// Transient failures are retried by default. See [RetryConfig] for the
// conditions, and [WithRetry] and [WithoutRetry] to change them.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
		Retry:      DefaultRetryConfig(),
	}
	for _, o := range opts {
		o.apply(c)
	}
	return c
}

func (c *Client) do(ctx context.Context, method, path string, reqBody, respBody any) error {
	var body []byte
	if reqBody != nil {
		var err error
		body, err = json.Marshal(reqBody)
		if err != nil {
			return err
		}
	}

	attempts := c.maxAttempts()
	for attempt := 1; ; attempt++ {
		var bodyReader io.Reader
		if reqBody != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
		if err != nil {
			return err
		}
		if reqBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)

		data, resp, err := c.attempt(req)

		if attempt < attempts && (err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300) && c.retryable(req, resp, err) {
			if delay, ok := c.retryDelay(attempt, resp); ok {
				if sleep(ctx, delay) == nil {
					continue
				}
			}
		}

		if err != nil {
			return err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			apiErr := &Error{StatusCode: resp.StatusCode}
			if err := json.Unmarshal(data, apiErr); err != nil {
				return &Error{
					StatusCode: resp.StatusCode,
					Message:    string(data),
				}
			}
			return apiErr
		}

		if respBody != nil && len(data) > 0 {
			return json.Unmarshal(data, respBody)
		}
		return nil
	}
}

// attempt performs a single request and reads its response body. It returns a
// nil response only when err is non-nil.
func (c *Client) attempt(req *http.Request) ([]byte, *http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	return data, resp, nil
}
