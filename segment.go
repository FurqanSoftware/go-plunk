package plunk

import (
	"context"
	"net/http"
)

// Segment represents a Plunk segment.
type Segment struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Filters         SegmentFilters `json:"filters"`
	TrackMembership bool           `json:"trackMembership"`
	MemberCount     int            `json:"memberCount"`
}

// SegmentFilters represents the filter conditions for a segment.
type SegmentFilters struct {
	Operator   string            `json:"operator"`
	Conditions []SegmentCondition `json:"conditions"`
}

// SegmentCondition represents a single filter condition within a segment.
type SegmentCondition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// CreateSegmentRequest is the request for [Client.CreateSegment].
type CreateSegmentRequest struct {
	Name            string         `json:"name"`
	Filters         SegmentFilters `json:"filters"`
	TrackMembership bool           `json:"trackMembership,omitempty"`
}

// CreateSegment creates a new segment. It requires a secret API key.
func (c *Client) CreateSegment(ctx context.Context, req *CreateSegmentRequest) (*Segment, error) {
	var resp Segment
	if err := c.do(ctx, http.MethodPost, "/segments", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListSegments retrieves all segments. It requires a secret API key.
func (c *Client) ListSegments(ctx context.Context) ([]Segment, error) {
	var resp []Segment
	if err := c.do(ctx, http.MethodGet, "/segments", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}
