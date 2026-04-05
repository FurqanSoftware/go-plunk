package plunk

import "context"

// SendRequest is the request for [Client.Send].
type SendRequest struct {
	To          []Address
	From        Address
	Subject     string
	Body        string
	Template    string
	Name        string
	Subscribed  bool
	Data        map[string]any
	Headers     map[string]string
	Reply       string
	Attachments []Attachment
}

// Attachment represents a file attachment for a transactional email. Content
// must be base64-encoded. A maximum of 10 attachments totaling 10 MB is
// allowed per request.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

// SendResponse is the response from [Client.Send].
type SendResponse struct {
	Emails    []SendEmail `json:"emails"`
	Timestamp string      `json:"timestamp"`
}

// SendEmail represents a sent email in the response.
type SendEmail struct {
	Contact SendContact `json:"contact"`
	Email   string      `json:"email"`
}

// SendContact represents a contact associated with a sent email.
type SendContact struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type sendRequest struct {
	To          toField           `json:"to"`
	From        Address           `json:"from"`
	Subject     string            `json:"subject,omitempty"`
	Body        string            `json:"body,omitempty"`
	Template    string            `json:"template,omitempty"`
	Name        string            `json:"name,omitempty"`
	Subscribed  bool              `json:"subscribed,omitempty"`
	Data        map[string]any    `json:"data,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Reply       string            `json:"reply,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
}

// Send sends a transactional email. It requires a secret API key.
func (c *Client) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
	jr := sendRequest{
		To:          toField(req.To),
		From:        req.From,
		Subject:     req.Subject,
		Body:        req.Body,
		Template:    req.Template,
		Name:        req.Name,
		Subscribed:  req.Subscribed,
		Data:        req.Data,
		Headers:     req.Headers,
		Reply:       req.Reply,
		Attachments: req.Attachments,
	}

	var resp struct {
		Success bool         `json:"success"`
		Data    SendResponse `json:"data"`
	}
	if err := c.do(ctx, "/v1/send", jr, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
