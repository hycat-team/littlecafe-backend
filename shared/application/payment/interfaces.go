// Package payment defines the outbound payment provider port.
package payment

import "context"

// Link is a provider payment link.
type Link struct {
	URL       string
	Reference string
}

// Provider creates payment links and verifies provider callbacks.
type Provider interface {
	CreateLink(ctx context.Context, orderID string, amount int64, description string) (Link, error)
	VerifyCallback(payload []byte, signature string) error
}
