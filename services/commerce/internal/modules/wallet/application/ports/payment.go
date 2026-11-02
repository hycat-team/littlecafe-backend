package ports

import "context"

// PaymentLink is a provider payment link.
type PaymentLink struct {
	URL       string
	Reference string
}

// PaymentProvider creates payment links and verifies provider callbacks.
type PaymentProvider interface {
	CreateLink(ctx context.Context, orderID string, amount int64, description string) (PaymentLink, error)
	VerifyCallback(payload []byte, signature string) error
}
