// Package oauth implements the ports.OAuthProvider contract for Google using
// only the standard library (no provider SDK).
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	identityconfig "github.com/hycat-team/littlecafe-backend/services/identity/internal/config"
	"github.com/hycat-team/littlecafe-backend/services/identity/internal/modules/identity/application/ports"
)

const (
	googleAuthEndpoint  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenEndpoint = "https://oauth2.googleapis.com/token"
	googleUserEndpoint  = "https://openidconnect.googleapis.com/v1/userinfo"
)

// GoogleProvider implements ports.OAuthProvider against Google.
type GoogleProvider struct {
	cfg    *identityconfig.Config
	client *http.Client
}

// NewGoogleProvider builds the Google adapter.
func NewGoogleProvider(cfg *identityconfig.Config) *GoogleProvider {
	return &GoogleProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthURL builds the consent URL with the configured redirect URI and scopes.
func (p *GoogleProvider) AuthURL(state string) string {
	q := url.Values{}
	q.Set("client_id", p.cfg.Google.WebClientID)
	q.Set("redirect_uri", p.cfg.Google.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	return googleAuthEndpoint + "?" + q.Encode()
}

// Exchange swaps the authorization code for the user's normalized identity.
func (p *GoogleProvider) Exchange(ctx context.Context, code string) (ports.OAuthIdentity, error) {
	token, err := p.exchangeCode(ctx, code)
	if err != nil {
		return ports.OAuthIdentity{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserEndpoint, nil)
	if err != nil {
		return ports.OAuthIdentity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := p.client.Do(req)
	if err != nil {
		return ports.OAuthIdentity{}, fmt.Errorf("oauth: userinfo request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ports.OAuthIdentity{}, fmt.Errorf("oauth: userinfo returned %d", resp.StatusCode)
	}
	var profile struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&profile); err != nil {
		return ports.OAuthIdentity{}, fmt.Errorf("oauth: decode userinfo: %w", err)
	}
	return ports.OAuthIdentity{ProviderUserID: profile.Sub, Email: profile.Email, Name: profile.Name}, nil
}

func (p *GoogleProvider) exchangeCode(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.cfg.Google.WebClientID)
	form.Set("client_secret", p.cfg.Google.WebClientSecret)
	form.Set("redirect_uri", p.cfg.Google.RedirectURI)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth: token endpoint returned %d", resp.StatusCode)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", fmt.Errorf("oauth: decode token response: %w", err)
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("oauth: empty access token")
	}
	return payload.AccessToken, nil
}

var _ ports.OAuthProvider = (*GoogleProvider)(nil)
