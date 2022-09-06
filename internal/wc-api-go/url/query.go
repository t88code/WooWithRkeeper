package url // import "WooWithRkeeper/internal/wc-api-go/url"

import (
	"WooWithRkeeper/internal/wc-api-go/request"
	"net/url"
)

// QueryEnricher uses package auth to enrich existing query parameters with Authentication Based ones
type QueryEnricher interface {
	GetEnrichedQuery(url string, query url.Values, req request.Request) url.Values
}
