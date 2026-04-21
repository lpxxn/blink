package mail

import "context"

// Message is a single outbound email. HTMLBody is optional; when empty, only
// the plain-text body is sent.
type Message struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
}

// Mailer sends verification / notification emails. Implementations MUST be
// safe for concurrent use.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}
