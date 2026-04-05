package plunk

import "context"

// TrackRequest is the request for [Client.Track]. Subscribed defaults to true
// on the server when nil.
type TrackRequest struct {
	Email      string         `json:"email"`
	Event      string         `json:"event"`
	Subscribed *bool          `json:"subscribed,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}

// TrackResponse is the response from [Client.Track].
type TrackResponse struct {
	Contact   string `json:"contact"`
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
}

// Track tracks an event for a contact. It accepts both secret and public API
// keys.
func (c *Client) Track(ctx context.Context, req *TrackRequest) (*TrackResponse, error) {
	var resp struct {
		Success bool          `json:"success"`
		Data    TrackResponse `json:"data"`
	}
	if err := c.do(ctx, "/v1/track", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
