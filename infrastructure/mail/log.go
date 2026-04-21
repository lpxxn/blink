package mail

import (
	"context"
	"log"

	domainmail "github.com/lpxxn/blink/domain/mail"
)

// LogMailer is a development / fallback Mailer that writes message metadata to
// the standard logger. It is also used when SMTP is not configured yet.
type LogMailer struct {
	Logger *log.Logger
}

func (m *LogMailer) Send(_ context.Context, msg domainmail.Message) error {
	l := m.Logger
	if l == nil {
		l = log.Default()
	}
	l.Printf("[mail] (log-only) to=%s subject=%q body=%q",
		maskEmail(msg.To), msg.Subject, msg.TextBody)
	return nil
}

// maskEmail turns alice@example.com into a***@example.com so that logs remain
// useful for debugging without leaking PII.
func maskEmail(e string) string {
	at := -1
	for i := 0; i < len(e); i++ {
		if e[i] == '@' {
			at = i
			break
		}
	}
	if at <= 0 {
		return e
	}
	local := e[:at]
	domain := e[at:]
	if len(local) == 1 {
		return local + "***" + domain
	}
	return string(local[0]) + "***" + domain
}
