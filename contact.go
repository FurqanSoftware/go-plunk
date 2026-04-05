package plunk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Contact represents a Plunk contact.
type Contact struct {
	ID         string         `json:"id"`
	Email      string         `json:"email"`
	Subscribed bool           `json:"subscribed"`
	Data       map[string]any `json:"data"`
	CreatedAt  string         `json:"createdAt"`
	UpdatedAt  string         `json:"updatedAt"`
}

// CreateContactRequest is the request for [Client.CreateContact].
type CreateContactRequest struct {
	Email      string         `json:"email"`
	Subscribed *bool          `json:"subscribed,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}

// CreateContactResponse is the response from [Client.CreateContact].
type CreateContactResponse struct {
	Contact
	Meta struct {
		IsNew    bool `json:"isNew"`
		IsUpdate bool `json:"isUpdate"`
	} `json:"_meta"`
}

// CreateContact creates a new contact or updates an existing one by email. It
// requires a secret API key.
func (c *Client) CreateContact(ctx context.Context, req *CreateContactRequest) (*CreateContactResponse, error) {
	var resp CreateContactResponse
	if err := c.do(ctx, http.MethodPost, "/contacts", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetContact retrieves a contact by ID. It requires a secret API key.
func (c *Client) GetContact(ctx context.Context, id string) (*Contact, error) {
	var resp Contact
	if err := c.do(ctx, http.MethodGet, "/contacts/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateContactRequest is the request for [Client.UpdateContact].
type UpdateContactRequest struct {
	Subscribed *bool          `json:"subscribed,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}

// UpdateContact updates an existing contact by ID. It requires a secret API
// key.
func (c *Client) UpdateContact(ctx context.Context, id string, req *UpdateContactRequest) (*Contact, error) {
	var resp Contact
	if err := c.do(ctx, http.MethodPatch, "/contacts/"+id, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteContact permanently deletes a contact by ID. It requires a secret API
// key.
func (c *Client) DeleteContact(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/contacts/"+id, nil, nil)
}

// ListContactsRequest is the request for [Client.ListContacts].
type ListContactsRequest struct {
	Limit      int
	Cursor     string
	Subscribed *bool
	Search     string
}

// ListContactsResponse is the response from [Client.ListContacts].
type ListContactsResponse struct {
	Contacts []Contact `json:"contacts"`
	Cursor   *string   `json:"cursor"`
	HasMore  bool      `json:"hasMore"`
	Total    int       `json:"total"`
}

// ListContacts retrieves a paginated list of contacts. It requires a secret API
// key.
func (c *Client) ListContacts(ctx context.Context, req *ListContactsRequest) (*ListContactsResponse, error) {
	q := url.Values{}
	if req != nil {
		if req.Limit > 0 {
			q.Set("limit", strconv.Itoa(req.Limit))
		}
		if req.Cursor != "" {
			q.Set("cursor", req.Cursor)
		}
		if req.Subscribed != nil {
			q.Set("subscribed", fmt.Sprintf("%t", *req.Subscribed))
		}
		if req.Search != "" {
			q.Set("search", req.Search)
		}
	}

	path := "/contacts"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var resp ListContactsResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
