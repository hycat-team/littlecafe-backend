package ports

import "context"

// OAuthIdentity is the normalized profile returned by an external identity
// provider.
type OAuthIdentity struct {
	ProviderUserID string
	Email          string
	Name           string
}

// OAuthProvider builds an authorization URL and exchanges the callback code.
type OAuthProvider interface {
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (OAuthIdentity, error)
}
