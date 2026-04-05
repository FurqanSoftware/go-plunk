package plunk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreateContact(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/contacts" {
			t.Errorf("path = %s, want /contacts", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["email"] != "user@example.com" {
			t.Errorf("email = %v, want user@example.com", req["email"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":         "c1",
			"email":      "user@example.com",
			"subscribed": true,
			"data":       map[string]any{"plan": "premium"},
			"createdAt":  "2025-01-01T00:00:00Z",
			"updatedAt":  "2025-01-01T00:00:00Z",
			"_meta":      map[string]any{"isNew": true, "isUpdate": false},
		})
	})

	resp, err := c.CreateContact(context.Background(), &CreateContactRequest{
		Email: "user@example.com",
		Data:  map[string]any{"plan": "premium"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "c1" {
		t.Errorf("id = %s, want c1", resp.ID)
	}
	if !resp.Meta.IsNew {
		t.Error("meta.isNew = false, want true")
	}
}

func TestGetContact(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/contacts/c1" {
			t.Errorf("path = %s, want /contacts/c1", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":         "c1",
			"email":      "user@example.com",
			"subscribed": true,
			"data":       map[string]any{},
			"createdAt":  "2025-01-01T00:00:00Z",
			"updatedAt":  "2025-01-01T00:00:00Z",
		})
	})

	resp, err := c.GetContact(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "c1" {
		t.Errorf("id = %s, want c1", resp.ID)
	}
	if resp.Email != "user@example.com" {
		t.Errorf("email = %s, want user@example.com", resp.Email)
	}
}

func TestUpdateContact(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/contacts/c1" {
			t.Errorf("path = %s, want /contacts/c1", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["subscribed"] != false {
			t.Errorf("subscribed = %v, want false", req["subscribed"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":         "c1",
			"email":      "user@example.com",
			"subscribed": false,
			"data":       map[string]any{},
			"createdAt":  "2025-01-01T00:00:00Z",
			"updatedAt":  "2025-01-02T00:00:00Z",
		})
	})

	sub := false
	resp, err := c.UpdateContact(context.Background(), "c1", &UpdateContactRequest{
		Subscribed: &sub,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Subscribed {
		t.Error("subscribed = true, want false")
	}
}

func TestDeleteContact(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/contacts/c1" {
			t.Errorf("path = %s, want /contacts/c1", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := c.DeleteContact(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestListContacts(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/contacts" {
			t.Errorf("path = %s, want /contacts", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("limit = %s, want 10", got)
		}
		if got := r.URL.Query().Get("search"); got != "user" {
			t.Errorf("search = %s, want user", got)
		}

		w.Header().Set("Content-Type", "application/json")
		cursor := "cur1"
		json.NewEncoder(w).Encode(map[string]any{
			"contacts": []any{
				map[string]any{
					"id":         "c1",
					"email":      "user@example.com",
					"subscribed": true,
					"data":       map[string]any{},
					"createdAt":  "2025-01-01T00:00:00Z",
					"updatedAt":  "2025-01-01T00:00:00Z",
				},
			},
			"cursor":  cursor,
			"hasMore": true,
			"total":   50,
		})
	})

	resp, err := c.ListContacts(context.Background(), &ListContactsRequest{
		Limit:  10,
		Search: "user",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Contacts) != 1 {
		t.Fatalf("len(contacts) = %d, want 1", len(resp.Contacts))
	}
	if resp.Contacts[0].ID != "c1" {
		t.Errorf("contacts[0].id = %s, want c1", resp.Contacts[0].ID)
	}
	if !resp.HasMore {
		t.Error("hasMore = false, want true")
	}
	if resp.Total != 50 {
		t.Errorf("total = %d, want 50", resp.Total)
	}
}
