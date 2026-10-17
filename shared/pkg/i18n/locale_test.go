package i18n

import (
	"context"
	"testing"
)

func TestNegotiate(t *testing.T) {
	negotiator := DefaultNegotiator()

	tests := []struct {
		name   string
		header string
		want   Locale
		wantOK bool
	}{
		{name: "empty falls back", header: "", want: VI, wantOK: false},
		{name: "exact vi", header: "vi", want: VI, wantOK: true},
		{name: "regional vi", header: "vi-VN", want: VI, wantOK: true},
		{name: "exact en", header: "en", want: EN, wantOK: true},
		{name: "regional en with quality", header: "en-US,en;q=0.9", want: EN, wantOK: true},
		{name: "prefers first supported", header: "fr-FR,en;q=0.8", want: EN, wantOK: true},
		{name: "unsupported only", header: "fr-FR", want: VI, wantOK: false},
		{name: "malformed", header: ";;;", want: VI, wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := negotiator.Negotiate(tc.header)
			if got != tc.want || ok != tc.wantOK {
				t.Fatalf("Negotiate(%q) = (%s, %v), want (%s, %v)", tc.header, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestParsePreference(t *testing.T) {
	negotiator := DefaultNegotiator()

	tests := []struct {
		tag    string
		want   Locale
		wantOK bool
	}{
		{tag: "", want: VI, wantOK: false},
		{tag: "en-GB", want: EN, wantOK: true},
		{tag: "vi", want: VI, wantOK: true},
		{tag: "fr", want: VI, wantOK: false},
	}

	for _, tc := range tests {
		got, ok := negotiator.Parse(tc.tag)
		if got != tc.want || ok != tc.wantOK {
			t.Fatalf("Parse(%q) = (%s, %v), want (%s, %v)", tc.tag, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestWithLocaleRoundTrip(t *testing.T) {
	ctx := WithLocale(context.Background(), EN)
	if got := FromContext(ctx); got != EN {
		t.Fatalf("FromContext = %s, want en", got)
	}
	if got := FromContext(context.Background()); got != VI {
		t.Fatalf("default FromContext = %s, want vi", got)
	}
	if got := FromContext(WithLocale(nil, "xx")); got != VI {
		t.Fatalf("unsupported locale = %s, want vi", got)
	}
}
