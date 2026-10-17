// Package oauth defines the external identity provider port.
package oauth

import "context"

// Identity is the normalized profile returned by a provider.
type Identity struct {
	ProviderUserID string
	Email          string
	Name           string
}

// Provider builds an authorization URL and exchanges the callback code.
type Provider interface {
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (Identity, error)
}
