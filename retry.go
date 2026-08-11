package plunk

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Default retry settings used by [New].
const (
	DefaultMaxAttempts   = 3
	DefaultMinRetryDelay = 500 * time.Millisecond
	DefaultMaxRetryDelay = 30 * time.Second
)

// RetryConfig configures how a [Client] retries failed requests.
//
// Requests are retried for transient failures: network errors and the status
// codes 408, 429 and 5xx. Because the Plunk API has no idempotency key,
// retries apply to non-idempotent requests such as [Client.Send] as well. A
// request that reached Plunk but whose response was lost in transit is
// therefore retried, and the email is sent twice. Set Retryable to narrow the
// conditions if duplicate sends are worse than lost ones.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts made, including the first
	// one. Values below 2 disable retries. If zero, [DefaultMaxAttempts] is
	// used.
	MaxAttempts int

	// MinDelay is the base delay of the exponential backoff, doubling with
	// every attempt. If zero, [DefaultMinRetryDelay] is used.
	MinDelay time.Duration

	// MaxDelay caps the backoff delay. If zero, [DefaultMaxRetryDelay] is
	// used.
	MaxDelay time.Duration

	// Retryable reports whether a failed attempt should be retried. If nil,
	// [DefaultRetryable] is used.
	//
	// A non-nil err means the request failed before a usable response was
	// obtained, in which case resp may be nil. The body of resp has already
	// been consumed and closed; it must not be read.
	Retryable func(req *http.Request, resp *http.Response, err error) bool
}

// DefaultRetryConfig returns the retry configuration used by [New].
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: DefaultMaxAttempts,
		MinDelay:    DefaultMinRetryDelay,
		MaxDelay:    DefaultMaxRetryDelay,
		Retryable:   DefaultRetryable,
	}
}

// WithRetry returns an [Option] that sets the retry configuration. Zero-valued
// fields fall back to their defaults.
func WithRetry(rc RetryConfig) Option {
	return OptionFunc(func(c *Client) {
		c.Retry = rc
	})
}

// WithoutRetry returns an [Option] that disables retries, so that every request
// is attempted exactly once.
func WithoutRetry() Option {
	return OptionFunc(func(c *Client) {
		c.Retry.MaxAttempts = 1
	})
}

// DefaultRetryable reports whether a failed attempt should be retried. It
// retries network errors, and responses with status 408, 429 or 5xx other than
// 501. It does not retry once the request context is done.
func DefaultRetryable(req *http.Request, resp *http.Response, err error) bool {
	if req.Context().Err() != nil {
		return false
	}
	if err != nil {
		return true
	}
	switch {
	case resp.StatusCode == http.StatusRequestTimeout,
		resp.StatusCode == http.StatusTooManyRequests:
		return true
	case resp.StatusCode == http.StatusNotImplemented:
		return false
	default:
		return resp.StatusCode >= 500
	}
}

func (c *Client) maxAttempts() int {
	if c.Retry.MaxAttempts < 1 {
		if c.Retry.MaxAttempts == 0 {
			return DefaultMaxAttempts
		}
		return 1
	}
	return c.Retry.MaxAttempts
}

func (c *Client) retryable(req *http.Request, resp *http.Response, err error) bool {
	if c.Retry.Retryable == nil {
		return DefaultRetryable(req, resp, err)
	}
	return c.Retry.Retryable(req, resp, err)
}

// retryDelay returns how long to wait before attempt+1. It reports false if the
// server asked for a longer delay than the client is willing to wait, in which
// case the request is not retried.
func (c *Client) retryDelay(attempt int, resp *http.Response) (time.Duration, bool) {
	min, max := c.Retry.MinDelay, c.Retry.MaxDelay
	if min <= 0 {
		min = DefaultMinRetryDelay
	}
	if max <= 0 {
		max = DefaultMaxRetryDelay
	}
	if max < min {
		max = min
	}

	if resp != nil {
		if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			if d > max {
				return 0, false
			}
			return d, true
		}
	}

	// Exponential backoff with equal jitter: half of the delay is fixed, the
	// other half random, so that concurrent clients spread out.
	d := min
	for i := 1; i < attempt && d < max; i++ {
		d *= 2
	}
	if d > max || d <= 0 {
		d = max
	}
	return d/2 + time.Duration(rand.Int63n(int64(d/2)+1)), true
}

// parseRetryAfter parses a Retry-After header value, given either as a number
// of seconds or as an HTTP date.
func parseRetryAfter(v string) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	t, err := http.ParseTime(v)
	if err != nil {
		return 0, false
	}
	d := time.Until(t)
	if d < 0 {
		d = 0
	}
	return d, true
}

// sleep waits for d, or until ctx is done.
func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
