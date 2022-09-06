package request

import (
	"net/url"
)

// Request ...
type Request struct {
	Method   string
	Endpoint string
	Values   url.Values
	Body     interface{}
}
