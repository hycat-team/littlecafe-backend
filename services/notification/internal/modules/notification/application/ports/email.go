package ports

import "context"

// EmailMessage is one outbound email.
type EmailMessage struct {
	To       []string
	Subject  string
	HTMLBody string
}

// EmailSender delivers email messages.
type EmailSender interface {
	Send(ctx context.Context, message EmailMessage) error
}
