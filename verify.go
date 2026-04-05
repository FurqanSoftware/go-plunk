package plunk

import "context"

// VerifyRequest is the request for [Client.Verify].
type VerifyRequest struct {
	Email string `json:"email"`
}

// VerifyResponse is the response from [Client.Verify].
type VerifyResponse struct {
	Email           string   `json:"email"`
	Valid           bool     `json:"valid"`
	IsDisposable    bool     `json:"isDisposable"`
	IsAlias         bool     `json:"isAlias"`
	IsTypo          bool     `json:"isTypo"`
	IsPlusAddressed bool     `json:"isPlusAddressed"`
	IsPersonalEmail bool     `json:"isPersonalEmail"`
	DomainExists    bool     `json:"domainExists"`
	HasWebsite      bool     `json:"hasWebsite"`
	HasMxRecords    bool     `json:"hasMxRecords"`
	Reasons         []string `json:"reasons"`
}

// Verify verifies an email address. It requires a secret API key.
func (c *Client) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	var resp struct {
		Success bool           `json:"success"`
		Data    VerifyResponse `json:"data"`
	}
	if err := c.do(ctx, "/v1/verify", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
