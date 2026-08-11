package plunk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// fastRetry returns a retry configuration with negligible delays, so that tests
// exercise the retry logic without sleeping.
func fastRetry(maxAttempts int) RetryConfig {
	return RetryConfig{
		MaxAttempts: maxAttempts,
		MinDelay:    time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
	}
}

func retryTestClient(t *testing.T, rc RetryConfig, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("sk_test", WithBaseURL(srv.URL), WithRetry(rc))
}

func TestRetryOnServerError(t *testing.T) {
	var calls int32
	c := retryTestClient(t, fastRetry(3), func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]any{"contact": "c1", "event": "ev1"},
		})
	})

	resp, err := c.Track(context.Background(), &TrackRequest{
		Email: "user@example.com",
		Event: "signup",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Contact != "c1" {
		t.Errorf("contact = %s, want c1", resp.Contact)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestRetryExhausted(t *testing.T) {
	var calls int32
	c := retryTestClient(t, fastRetry(3), func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    503,
			"message": "Service Unavailable",
		})
	})

	_, err := c.Track(context.Background(), &TrackRequest{Email: "user@example.com", Event: "signup"})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %v", err)
	}
	if apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status code = %d, want 503", apiErr.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestNoRetryOnClientError(t *testing.T) {
	var calls int32
	c := retryTestClient(t, fastRetry(3), func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    401,
			"message": "Invalid API key",
		})
	})

	_, err := c.Track(context.Background(), &TrackRequest{Email: "user@example.com", Event: "signup"})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestRetryReplaysBody(t *testing.T) {
	var calls int32
	var bodies []string
	c := retryTestClient(t, fastRetry(2), func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)
		bodies = append(bodies, req["subject"].(string))

		if atomic.AddInt32(&calls, 1) < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]any{"timestamp": "2025-01-01T00:00:00Z"},
		})
	})

	_, err := c.Send(context.Background(), &SendRequest{
		To:      []Address{Addr("user@example.com")},
		From:    Address{Email: "hello@acme.com"},
		Subject: "Hello",
		Body:    "<p>Hi</p>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(bodies) != 2 {
		t.Fatalf("len(bodies) = %d, want 2", len(bodies))
	}
	for i, subject := range bodies {
		if subject != "Hello" {
			t.Errorf("bodies[%d] = %q, want Hello", i, subject)
		}
	}
}

func TestRetryDisabled(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New("sk_test", WithBaseURL(srv.URL), WithoutRetry())
	_, err := c.Track(context.Background(), &TrackRequest{Email: "user@example.com", Event: "signup"})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestRetryAfterHeader(t *testing.T) {
	var calls int32
	start := time.Now()
	c := retryTestClient(t, RetryConfig{MaxAttempts: 2, MinDelay: time.Hour, MaxDelay: time.Minute}, func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"success": true, "data": map[string]any{}})
	})

	_, err := c.Track(context.Background(), &TrackRequest{Email: "user@example.com", Event: "signup"})
	if err != nil {
		t.Fatal(err)
	}
	// Retry-After takes precedence over the (one hour) exponential backoff.
	if elapsed := time.Since(start); elapsed < time.Second || elapsed > 30*time.Second {
		t.Errorf("elapsed = %s, want about 1s", elapsed)
	}
}

func TestRetryAfterTooLong(t *testing.T) {
	var calls int32
	c := retryTestClient(t, RetryConfig{MaxAttempts: 3, MinDelay: time.Millisecond, MaxDelay: time.Second}, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := c.Track(context.Background(), &TrackRequest{Email: "user@example.com", Event: "signup"})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %v", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status code = %d, want 429", apiErr.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestRetryOnNetworkError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 2 {
			// Close the connection without writing a response.
			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Error("response writer is not a hijacker")
				return
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Error(err)
				return
			}
			conn.Close()
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data":    map[string]any{"contact": "c1"},
		})
	}))
	defer srv.Close()

	c := New("sk_test", WithBaseURL(srv.URL), WithRetry(fastRetry(3)))
	resp, err := c.Track(context.Background(), &TrackRequest{Email: "user@example.com", Event: "signup"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Contact != "c1" {
		t.Errorf("contact = %s, want c1", resp.Contact)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestNoRetryAfterContextCanceled(t *testing.T) {
	var calls int32
	c := retryTestClient(t, RetryConfig{MaxAttempts: 3, MinDelay: time.Minute, MaxDelay: time.Minute}, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := c.Track(ctx, &TrackRequest{Email: "user@example.com", Event: "signup"})
	if err == nil {
		t.Fatal("expected error")
	}
	// The backoff must not outlive the context.
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Errorf("elapsed = %s, want the backoff to be cut short", elapsed)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}

func TestCustomRetryable(t *testing.T) {
	var calls int32
	rc := fastRetry(3)
	// Retry only idempotent requests.
	rc.Retryable = func(req *http.Request, resp *http.Response, err error) bool {
		if req.Method == http.MethodPost {
			return false
		}
		return DefaultRetryable(req, resp, err)
	}
	c := retryTestClient(t, rc, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := c.Track(context.Background(), &TrackRequest{Email: "user@example.com", Event: "signup"})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}

	atomic.StoreInt32(&calls, 0)
	_, err = c.GetContact(context.Background(), "c1")
	if err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestParseRetryAfter(t *testing.T) {
	if d, ok := parseRetryAfter("5"); !ok || d != 5*time.Second {
		t.Errorf("parseRetryAfter(5) = %s, %t, want 5s, true", d, ok)
	}
	if _, ok := parseRetryAfter(""); ok {
		t.Error("parseRetryAfter() = true, want false")
	}
	if _, ok := parseRetryAfter("later"); ok {
		t.Error("parseRetryAfter(later) = true, want false")
	}
	if _, ok := parseRetryAfter("-1"); ok {
		t.Error("parseRetryAfter(-1) = true, want false")
	}
	// An HTTP date in the past yields no delay, but is still honored.
	if d, ok := parseRetryAfter("Wed, 01 Jan 2020 00:00:00 GMT"); !ok || d != 0 {
		t.Errorf("parseRetryAfter(past date) = %s, %t, want 0s, true", d, ok)
	}
	if d, ok := parseRetryAfter(time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)); !ok || d < 59*time.Minute {
		t.Errorf("parseRetryAfter(future date) = %s, %t, want about 1h, true", d, ok)
	}
}

func TestRetryDelayBackoff(t *testing.T) {
	c := New("sk_test", WithRetry(RetryConfig{MaxAttempts: 10, MinDelay: time.Second, MaxDelay: 8 * time.Second}))

	for _, tt := range []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{1, 500 * time.Millisecond, time.Second},
		{2, time.Second, 2 * time.Second},
		{3, 2 * time.Second, 4 * time.Second},
		{4, 4 * time.Second, 8 * time.Second},
		{5, 4 * time.Second, 8 * time.Second},
		{50, 4 * time.Second, 8 * time.Second},
	} {
		d, ok := c.retryDelay(tt.attempt, nil)
		if !ok {
			t.Fatalf("retryDelay(%d) = false, want true", tt.attempt)
		}
		if d < tt.min || d > tt.max {
			t.Errorf("retryDelay(%d) = %s, want between %s and %s", tt.attempt, d, tt.min, tt.max)
		}
	}
}

func TestDefaultRetryable(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

	for _, tt := range []struct {
		code int
		want bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusNotFound, false},
		{http.StatusRequestTimeout, true},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusNotImplemented, false},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	} {
		if got := DefaultRetryable(req, &http.Response{StatusCode: tt.code}, nil); got != tt.want {
			t.Errorf("DefaultRetryable(%d) = %t, want %t", tt.code, got, tt.want)
		}
	}

	if !DefaultRetryable(req, nil, errors.New("connection reset")) {
		t.Error("DefaultRetryable(network error) = false, want true")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	canceled, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
	if DefaultRetryable(canceled, nil, errors.New("connection reset")) {
		t.Error("DefaultRetryable(canceled context) = true, want false")
	}
}
