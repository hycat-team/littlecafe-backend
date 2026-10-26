package oauth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	identityconfig "github.com/hycat-team/littlecafe-backend/services/identity/internal/config"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newProvider(handler roundTripFunc) *GoogleProvider {
	cfg := &identityconfig.Config{}
	cfg.Google.WebClientID = "client-id"
	cfg.Google.WebClientSecret = "secret"
	cfg.Google.RedirectURI = "https://app.test/callback"
	provider := NewGoogleProvider(cfg)
	provider.client = &http.Client{Transport: handler}
	return provider
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

func TestAuthURL(t *testing.T) {
	url := newProvider(nil).AuthURL("state-1")
	for _, want := range []string{
		"client_id=client-id",
		"redirect_uri=https%3A%2F%2Fapp.test%2Fcallback",
		"response_type=code",
		"state=state-1",
		"scope=openid+email+profile",
	} {
		if !strings.Contains(url, want) {
			t.Fatalf("AuthURL = %q, missing %q", url, want)
		}
	}
}

func TestExchangeSuccess(t *testing.T) {
	provider := newProvider(func(r *http.Request) (*http.Response, error) {
		switch {
		case strings.Contains(r.URL.String(), "token"):
			return jsonResponse(http.StatusOK, `{"access_token":"access-1"}`), nil
		case strings.Contains(r.URL.String(), "userinfo"):
			if r.Header.Get("Authorization") != "Bearer access-1" {
				return nil, errors.New("missing bearer token")
			}
			return jsonResponse(http.StatusOK, `{"sub":"google-1","email":"user@example.com","name":"User"}`), nil
		default:
			return nil, errors.New("unexpected URL " + r.URL.String())
		}
	})

	identity, err := provider.Exchange(context.Background(), "code-1")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if identity.ProviderUserID != "google-1" || identity.Email != "user@example.com" || identity.Name != "User" {
		t.Fatalf("identity = %+v", identity)
	}
}

func TestExchangeTokenFailures(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		err      error
	}{
		{name: "transport", err: errors.New("network down")},
		{name: "status", response: jsonResponse(http.StatusBadGateway, "bad")},
		{name: "invalid json", response: jsonResponse(http.StatusOK, "not-json")},
		{name: "empty token", response: jsonResponse(http.StatusOK, `{"access_token":""}`)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := newProvider(func(r *http.Request) (*http.Response, error) {
				return tc.response, tc.err
			})
			if _, err := provider.Exchange(context.Background(), "code"); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestExchangeUserInfoFailures(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		err      error
	}{
		{name: "transport", err: errors.New("network down")},
		{name: "status", response: jsonResponse(http.StatusUnauthorized, "bad")},
		{name: "invalid json", response: jsonResponse(http.StatusOK, "not-json")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			provider := newProvider(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return jsonResponse(http.StatusOK, `{"access_token":"access-1"}`), nil
				}
				return tc.response, tc.err
			})
			if _, err := provider.Exchange(context.Background(), "code"); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
