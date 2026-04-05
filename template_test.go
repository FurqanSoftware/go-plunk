package plunk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreateTemplate(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/templates" {
			t.Errorf("path = %s, want /templates", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["name"] != "Welcome" {
			t.Errorf("name = %v, want Welcome", req["name"])
		}
		if req["type"] != "TRANSACTIONAL" {
			t.Errorf("type = %v, want TRANSACTIONAL", req["type"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":        "t1",
			"name":      "Welcome",
			"subject":   "Welcome!",
			"body":      "<h1>Hi</h1>",
			"type":      "TRANSACTIONAL",
			"createdAt": "2025-01-01T00:00:00Z",
		})
	})

	resp, err := c.CreateTemplate(context.Background(), &CreateTemplateRequest{
		Name:    "Welcome",
		Subject: "Welcome!",
		Body:    "<h1>Hi</h1>",
		Type:    TemplateTransactional,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "t1" {
		t.Errorf("id = %s, want t1", resp.ID)
	}
	if resp.Type != TemplateTransactional {
		t.Errorf("type = %s, want TRANSACTIONAL", resp.Type)
	}
}

func TestListTemplates(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/templates" {
			t.Errorf("path = %s, want /templates", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Errorf("limit = %s, want 5", got)
		}
		if got := r.URL.Query().Get("type"); got != "MARKETING" {
			t.Errorf("type = %s, want MARKETING", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"templates": []any{
				map[string]any{
					"id":        "t1",
					"name":      "Promo",
					"subject":   "Sale!",
					"body":      "<p>50% off</p>",
					"type":      "MARKETING",
					"createdAt": "2025-01-01T00:00:00Z",
				},
			},
			"total":      10,
			"page":       1,
			"pageSize":   5,
			"totalPages": 2,
		})
	})

	resp, err := c.ListTemplates(context.Background(), &ListTemplatesRequest{
		Limit: 5,
		Type:  TemplateMarketing,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Templates) != 1 {
		t.Fatalf("len(templates) = %d, want 1", len(resp.Templates))
	}
	if resp.Templates[0].ID != "t1" {
		t.Errorf("templates[0].id = %s, want t1", resp.Templates[0].ID)
	}
	if resp.Total != 10 {
		t.Errorf("total = %d, want 10", resp.Total)
	}
	if resp.TotalPages != 2 {
		t.Errorf("totalPages = %d, want 2", resp.TotalPages)
	}
}
