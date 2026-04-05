package plunk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("sk_test", WithBaseURL(srv.URL))
}

func TestSend(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/send" {
			t.Errorf("path = %s, want /v1/send", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk_test" {
			t.Errorf("authorization = %s, want Bearer sk_test", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content-type = %s, want application/json", got)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["to"] != "user@example.com" {
			t.Errorf("to = %v, want user@example.com", req["to"])
		}
		if req["subject"] != "Hello" {
			t.Errorf("subject = %v, want Hello", req["subject"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"emails": []any{
					map[string]any{
						"contact": map[string]any{"id": "c1", "email": "user@example.com"},
						"email":   "e1",
					},
				},
				"timestamp": "2025-01-01T00:00:00Z",
			},
		})
	})

	resp, err := c.Send(context.Background(), &SendRequest{
		To:      []Address{Addr("user@example.com")},
		From:    Address{Name: "Acme", Email: "hello@acme.com"},
		Subject: "Hello",
		Body:    "<p>Hi</p>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Emails) != 1 {
		t.Fatalf("len(emails) = %d, want 1", len(resp.Emails))
	}
	if resp.Emails[0].Contact.ID != "c1" {
		t.Errorf("contact id = %s, want c1", resp.Emails[0].Contact.ID)
	}
	if resp.Timestamp != "2025-01-01T00:00:00Z" {
		t.Errorf("timestamp = %s, want 2025-01-01T00:00:00Z", resp.Timestamp)
	}
}

func TestTrack(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/track" {
			t.Errorf("path = %s, want /v1/track", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["email"] != "user@example.com" {
			t.Errorf("email = %v, want user@example.com", req["email"])
		}
		if req["event"] != "signup" {
			t.Errorf("event = %v, want signup", req["event"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"contact":   "c1",
				"event":     "ev1",
				"timestamp": "2025-01-01T00:00:00Z",
			},
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
	if resp.Event != "ev1" {
		t.Errorf("event = %s, want ev1", resp.Event)
	}
}

func TestVerify(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/verify" {
			t.Errorf("path = %s, want /v1/verify", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"email":           "user@example.com",
				"valid":           true,
				"isDisposable":    false,
				"isAlias":         false,
				"isTypo":          false,
				"isPlusAddressed": false,
				"isPersonalEmail": true,
				"domainExists":    true,
				"hasWebsite":      true,
				"hasMxRecords":    true,
				"reasons":        []string{"Email appears to be valid"},
			},
		})
	})

	resp, err := c.Verify(context.Background(), &VerifyRequest{
		Email: "user@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Valid {
		t.Error("valid = false, want true")
	}
	if !resp.IsPersonalEmail {
		t.Error("isPersonalEmail = false, want true")
	}
	if len(resp.Reasons) != 1 {
		t.Fatalf("len(reasons) = %d, want 1", len(resp.Reasons))
	}
}

func TestAPIError(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    401,
			"error":   "Unauthorized",
			"message": "Invalid API key",
			"time":    1234567890,
		})
	})

	_, err := c.Track(context.Background(), &TrackRequest{
		Email: "user@example.com",
		Event: "signup",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("status code = %d, want 401", apiErr.StatusCode)
	}
	if apiErr.Message != "Invalid API key" {
		t.Errorf("message = %s, want Invalid API key", apiErr.Message)
	}
}
