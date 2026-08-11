# go-plunk

Go client for the [Plunk](https://useplunk.com) API.

## Install

```sh
go get github.com/FurqanSoftware/go-plunk
```

## Usage

```go
import "github.com/FurqanSoftware/go-plunk"
```

### Create a Client

```go
client := plunk.New("sk_...")
```

Options can be passed to customize the client:

```go
client := plunk.New("sk_...",
    plunk.WithHTTPClient(customHTTPClient),
    plunk.WithBaseURL("https://custom-api.example.com"),
)
```

### Retries

Transient failures are retried automatically. By default a request is attempted
up to 3 times, with an exponential backoff starting at 500ms and capped at 30s,
jittered to spread out concurrent clients. A `Retry-After` header takes
precedence over the backoff; if it asks for a longer wait than the cap, the
request is not retried.

Network errors and the status codes 408, 429 and 5xx (other than 501) are
retried:

```go
client := plunk.New("sk_...", plunk.WithRetry(plunk.RetryConfig{
    MaxAttempts: 5,
    MinDelay:    time.Second,
    MaxDelay:    time.Minute,
}))

// Or turn retries off entirely.
client := plunk.New("sk_...", plunk.WithoutRetry())
```

The Plunk API has no idempotency key, so retries apply to non-idempotent
requests such as `Send` as well. A request that reached Plunk but whose
response was lost in transit is retried, and the email is sent twice. If a
duplicate send is worse than a lost one, narrow the conditions with
`Retryable`:

```go
client := plunk.New("sk_...", plunk.WithRetry(plunk.RetryConfig{
    Retryable: func(req *http.Request, resp *http.Response, err error) bool {
        // Retry only when the server explicitly did not process the request.
        if req.Method == http.MethodPost {
            return err == nil && resp.StatusCode == http.StatusTooManyRequests
        }
        return plunk.DefaultRetryable(req, resp, err)
    },
}))
```

### Send Transactional Email

```go
resp, err := client.Send(ctx, &plunk.SendRequest{
    To:      []plunk.Address{plunk.Addr("recipient@example.com")},
    From:    plunk.Address{Name: "Acme", Email: "hello@acme.com"},
    Subject: "Welcome!",
    Body:    "<h1>Hello</h1>",
})
```

### Track Event

```go
resp, err := client.Track(ctx, &plunk.TrackRequest{
    Email: "user@example.com",
    Event: "signup",
})
```

### Verify Email

```go
resp, err := client.Verify(ctx, &plunk.VerifyRequest{
    Email: "user@example.com",
})
fmt.Println(resp.Valid)
```

### Contacts

```go
// Create or update a contact
contact, err := client.CreateContact(ctx, &plunk.CreateContactRequest{
    Email: "user@example.com",
    Data:  map[string]any{"plan": "premium"},
})

// Get a contact
contact, err := client.GetContact(ctx, "contact_id")

// Update a contact
contact, err := client.UpdateContact(ctx, "contact_id", &plunk.UpdateContactRequest{
    Data: map[string]any{"plan": "enterprise"},
})

// Delete a contact
err := client.DeleteContact(ctx, "contact_id")

// List contacts
list, err := client.ListContacts(ctx, &plunk.ListContactsRequest{
    Limit:  20,
    Search: "user@example.com",
})
```

### Templates

```go
// Create a template
tmpl, err := client.CreateTemplate(ctx, &plunk.CreateTemplateRequest{
    Name:    "Welcome",
    Subject: "Welcome!",
    Body:    "<h1>Hello</h1>",
    Type:    plunk.TemplateTransactional,
})

// List templates
list, err := client.ListTemplates(ctx, &plunk.ListTemplatesRequest{
    Limit: 10,
    Type:  plunk.TemplateMarketing,
})
```

### Campaigns

```go
// Create a campaign
campaign, err := client.CreateCampaign(ctx, &plunk.CreateCampaignRequest{
    Name:         "Launch",
    Subject:      "We're live!",
    Body:         "<h1>Hello</h1>",
    From:         "hello@acme.com",
    AudienceType: plunk.AudienceAll,
})

// List campaigns
list, err := client.ListCampaigns(ctx, &plunk.ListCampaignsRequest{
    Status: plunk.CampaignDraft,
})

// Send a campaign immediately
err := client.SendCampaign(ctx, "campaign_id", &plunk.SendCampaignRequest{})

// Schedule a campaign
scheduled := "2025-06-01T10:00:00Z"
err := client.SendCampaign(ctx, "campaign_id", &plunk.SendCampaignRequest{
    ScheduledFor: &scheduled,
})
```

### Segments

```go
// Create a segment
segment, err := client.CreateSegment(ctx, &plunk.CreateSegmentRequest{
    Name: "Premium Users",
    Filters: plunk.SegmentFilters{
        Operator: "AND",
        Conditions: []plunk.SegmentCondition{
            {Field: "data.plan", Operator: "equals", Value: "premium"},
        },
    },
    TrackMembership: true,
})

// List segments
segments, err := client.ListSegments(ctx)
```

### Error Handling

API errors are returned as `*plunk.Error` and can be inspected with `errors.As`:

```go
var apiErr *plunk.Error
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Message)
}
```
