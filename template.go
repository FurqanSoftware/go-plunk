package plunk

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Template type constants.
const (
	TemplateTransactional = "TRANSACTIONAL"
	TemplateMarketing     = "MARKETING"
)

// Template represents a Plunk email template.
type Template struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
}

// CreateTemplateRequest is the request for [Client.CreateTemplate].
type CreateTemplateRequest struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Type    string `json:"type"`
}

// CreateTemplate creates a new email template. It requires a secret API key.
func (c *Client) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*Template, error) {
	var resp Template
	if err := c.do(ctx, http.MethodPost, "/templates", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListTemplatesRequest is the request for [Client.ListTemplates].
type ListTemplatesRequest struct {
	Limit  int
	Cursor string
	Type   string
	Search string
}

// ListTemplatesResponse is the response from [Client.ListTemplates].
type ListTemplatesResponse struct {
	Templates  []Template `json:"templates"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"pageSize"`
	TotalPages int        `json:"totalPages"`
}

// ListTemplates retrieves a paginated list of templates. It requires a secret
// API key.
func (c *Client) ListTemplates(ctx context.Context, req *ListTemplatesRequest) (*ListTemplatesResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Cursor != "" {
			q.Set("cursor", req.Cursor)
		}
		if req.Type != "" {
			q.Set("type", req.Type)
		}
		if req.Search != "" {
			q.Set("search", req.Search)
		}
	}

	path := "/templates"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var resp ListTemplatesResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
