package plunk

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Campaign audience type constants.
const (
	AudienceAll      = "ALL"
	AudienceSegment  = "SEGMENT"
	AudienceFiltered = "FILTERED"
)

// Campaign status constants.
const (
	CampaignDraft     = "DRAFT"
	CampaignScheduled = "SCHEDULED"
	CampaignSending   = "SENDING"
	CampaignSent      = "SENT"
)

// Campaign represents a Plunk campaign.
type Campaign struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Subject     string  `json:"subject"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	ScheduledAt *string `json:"scheduledAt"`
}

// CreateCampaignRequest is the request for [Client.CreateCampaign].
type CreateCampaignRequest struct {
	Name           string         `json:"name"`
	Subject        string         `json:"subject"`
	Body           string         `json:"body"`
	From           string         `json:"from"`
	AudienceType   string         `json:"audienceType"`
	Description    string         `json:"description,omitempty"`
	FromName       string         `json:"fromName,omitempty"`
	ReplyTo        string         `json:"replyTo,omitempty"`
	SegmentID      string         `json:"segmentId,omitempty"`
	AudienceFilter map[string]any `json:"audienceFilter,omitempty"`
}

// CreateCampaign creates a new campaign. It requires a secret API key.
func (c *Client) CreateCampaign(ctx context.Context, req *CreateCampaignRequest) (*Campaign, error) {
	var resp struct {
		Success bool     `json:"success"`
		Data    Campaign `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost, "/campaigns", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ListCampaignsRequest is the request for [Client.ListCampaigns].
type ListCampaignsRequest struct {
	Limit  int
	Cursor string
	Status string
}

// ListCampaignsResponse is the response from [Client.ListCampaigns].
type ListCampaignsResponse struct {
	Campaigns  []Campaign `json:"campaigns"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}

// ListCampaigns retrieves a paginated list of campaigns. It requires a secret
// API key.
func (c *Client) ListCampaigns(ctx context.Context, req *ListCampaignsRequest) (*ListCampaignsResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Cursor != "" {
			q.Set("cursor", req.Cursor)
		}
		if req.Status != "" {
			q.Set("status", req.Status)
		}
	}

	path := "/campaigns"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var resp ListCampaignsResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendCampaignRequest is the request for [Client.SendCampaign].
type SendCampaignRequest struct {
	ScheduledFor *string `json:"scheduledFor"`
}

// SendCampaign sends or schedules a campaign. Set ScheduledFor to an ISO 8601
// timestamp to schedule, or leave it nil to send immediately. It requires a
// secret API key.
func (c *Client) SendCampaign(ctx context.Context, id string, req *SendCampaignRequest) error {
	return c.do(ctx, http.MethodPost, "/campaigns/"+id+"/send", req, nil)
}
