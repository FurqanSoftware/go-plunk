package plunk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreateCampaign(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/campaigns" {
			t.Errorf("path = %s, want /campaigns", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["name"] != "Launch" {
			t.Errorf("name = %v, want Launch", req["name"])
		}
		if req["audienceType"] != "ALL" {
			t.Errorf("audienceType = %v, want ALL", req["audienceType"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"data": map[string]any{
				"id":          "camp1",
				"name":        "Launch",
				"subject":     "We're live!",
				"type":        "ALL",
				"status":      "DRAFT",
				"scheduledAt": nil,
			},
		})
	})

	resp, err := c.CreateCampaign(context.Background(), &CreateCampaignRequest{
		Name:         "Launch",
		Subject:      "We're live!",
		Body:         "<h1>Hello</h1>",
		From:         "hello@acme.com",
		AudienceType: AudienceAll,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "camp1" {
		t.Errorf("id = %s, want camp1", resp.ID)
	}
	if resp.Status != CampaignDraft {
		t.Errorf("status = %s, want DRAFT", resp.Status)
	}
}

func TestListCampaigns(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/campaigns" {
			t.Errorf("path = %s, want /campaigns", r.URL.Path)
		}
		if got := r.URL.Query().Get("status"); got != "SENT" {
			t.Errorf("status = %s, want SENT", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"campaigns": []any{
				map[string]any{
					"id":          "camp1",
					"name":        "Launch",
					"subject":     "We're live!",
					"type":        "ALL",
					"status":      "SENT",
					"scheduledAt": nil,
				},
			},
			"total":      1,
			"page":       1,
			"pageSize":   20,
			"totalPages": 1,
		})
	})

	resp, err := c.ListCampaigns(context.Background(), &ListCampaignsRequest{
		Status: CampaignSent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Campaigns) != 1 {
		t.Fatalf("len(campaigns) = %d, want 1", len(resp.Campaigns))
	}
	if resp.Campaigns[0].Status != CampaignSent {
		t.Errorf("status = %s, want SENT", resp.Campaigns[0].Status)
	}
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
}

func TestSendCampaign(t *testing.T) {
	t.Run("immediate", func(t *testing.T) {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if r.URL.Path != "/campaigns/camp1/send" {
				t.Errorf("path = %s, want /campaigns/camp1/send", r.URL.Path)
			}

			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			json.Unmarshal(body, &req)
			if req["scheduledFor"] != nil {
				t.Errorf("scheduledFor = %v, want nil", req["scheduledFor"])
			}

			w.WriteHeader(http.StatusOK)
		})

		err := c.SendCampaign(context.Background(), "camp1", &SendCampaignRequest{})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("scheduled", func(t *testing.T) {
		c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			json.Unmarshal(body, &req)
			if req["scheduledFor"] != "2025-06-01T10:00:00Z" {
				t.Errorf("scheduledFor = %v, want 2025-06-01T10:00:00Z", req["scheduledFor"])
			}

			w.WriteHeader(http.StatusOK)
		})

		scheduled := "2025-06-01T10:00:00Z"
		err := c.SendCampaign(context.Background(), "camp1", &SendCampaignRequest{
			ScheduledFor: &scheduled,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
