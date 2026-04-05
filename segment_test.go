package plunk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreateSegment(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/segments" {
			t.Errorf("path = %s, want /segments", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		json.Unmarshal(body, &req)
		if req["name"] != "Premium Users" {
			t.Errorf("name = %v, want Premium Users", req["name"])
		}
		if req["trackMembership"] != true {
			t.Errorf("trackMembership = %v, want true", req["trackMembership"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"id":   "seg1",
			"name": "Premium Users",
			"filters": map[string]any{
				"operator": "AND",
				"conditions": []any{
					map[string]any{
						"field":    "data.plan",
						"operator": "equals",
						"value":    "premium",
					},
				},
			},
			"trackMembership": true,
			"memberCount":     42,
		})
	})

	resp, err := c.CreateSegment(context.Background(), &CreateSegmentRequest{
		Name: "Premium Users",
		Filters: SegmentFilters{
			Operator: "AND",
			Conditions: []SegmentCondition{
				{Field: "data.plan", Operator: "equals", Value: "premium"},
			},
		},
		TrackMembership: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "seg1" {
		t.Errorf("id = %s, want seg1", resp.ID)
	}
	if resp.MemberCount != 42 {
		t.Errorf("memberCount = %d, want 42", resp.MemberCount)
	}
	if len(resp.Filters.Conditions) != 1 {
		t.Fatalf("len(conditions) = %d, want 1", len(resp.Filters.Conditions))
	}
	if resp.Filters.Conditions[0].Field != "data.plan" {
		t.Errorf("conditions[0].field = %s, want data.plan", resp.Filters.Conditions[0].Field)
	}
}

func TestListSegments(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/segments" {
			t.Errorf("path = %s, want /segments", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"id":              "seg1",
				"name":            "Premium Users",
				"filters":         map[string]any{},
				"trackMembership": true,
				"memberCount":     42,
			},
			{
				"id":              "seg2",
				"name":            "Free Users",
				"filters":         map[string]any{},
				"trackMembership": false,
				"memberCount":     100,
			},
		})
	})

	segments, err := c.ListSegments(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	if segments[0].ID != "seg1" {
		t.Errorf("segments[0].id = %s, want seg1", segments[0].ID)
	}
	if segments[1].MemberCount != 100 {
		t.Errorf("segments[1].memberCount = %d, want 100", segments[1].MemberCount)
	}
}
