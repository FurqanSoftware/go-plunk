package plunk

import "fmt"

// Error represents an error response from the Plunk API. It can be inspected
// using [errors.As].
type Error struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"code"`
	Err        string `json:"error"`
	Message    string `json:"message"`
	Time       int64  `json:"time"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("plunk: %s (code %d)", e.Message, e.Code)
}
