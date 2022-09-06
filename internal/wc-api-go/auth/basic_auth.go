package auth // import "WooWithRkeeper/internal/wc-api-go/auth"

import (
	"WooWithRkeeper/internal/wc-api-go/options"
	"net/url"
)

// BasicAuthentication structure stores all required parameter values
type BasicAuthentication struct{}

// GetEnrichedQuery method might get Parameters Enriched using Options
func (b *BasicAuthentication) GetEnrichedQuery(p url.Values, o options.Basic) url.Values {
	if true == o.Options.QueryStringAuth {
		p.Set("consumer_key", o.Key)
		p.Set("consumer_secret", o.Secret)
	}
	return p
}
