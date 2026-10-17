// Package email defines the outbound email port.
package email

import "context"

// Message is one outbound email.
type Message struct {
	To       []string
	Subject  string
	HTMLBody string
}

// Sender delivers email messages.
type Sender interface {
	Send(ctx context.Context, message Message) error
}
