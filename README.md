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

### Error Handling

API errors are returned as `*plunk.Error` and can be inspected with `errors.As`:

```go
var apiErr *plunk.Error
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Message)
}
```
