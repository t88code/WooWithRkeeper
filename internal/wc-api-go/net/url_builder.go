package net // import "WooWithRkeeper/internal/wc-api-go/net"

import (
	"WooWithRkeeper/internal/wc-api-go/request"
)

// URLBuilder interface
type URLBuilder interface {
	GetURL(req request.Request) string
}
