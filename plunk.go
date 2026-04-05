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
)

const defaultBaseURL = "https://next-api.useplunk.com"

// Option configures a [Client].
type Option interface {
	apply(*Client)
}

// OptionFunc is an adapter to allow the use of ordinary functions as [Option]s.
type OptionFunc func(*Client)

func (f OptionFunc) apply(c *Client) { f(c) }

// WithBaseURL returns an [Option] that sets the base URL of the API.
func WithBaseURL(url string) Option {
	return OptionFunc(func(c *Client) {
		c.BaseURL = url
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
}

// New creates a new Plunk API client with the given API key and options.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
	}
	for _, o := range opts {
		o.apply(c)
	}
	return c
}

func (c *Client) do(ctx context.Context, path string, reqBody, respBody any) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
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

	return json.Unmarshal(data, respBody)
}
