package mail

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gomail "github.com/wneessen/go-mail"

	domainmail "github.com/lpxxn/blink/domain/mail"
)

// SMTPConfig is the runtime configuration required to dial an SMTP server.
// It is produced by infrastructure/mail.ConfigurableMailer out of the
// domain/settings repository.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	Security string // "starttls" | "ssl" | "plain"
}

var ErrInvalidSMTPConfig = errors.New("mail: invalid SMTP config")

func (c SMTPConfig) validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("%w: host required", ErrInvalidSMTPConfig)
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("%w: port out of range", ErrInvalidSMTPConfig)
	}
	if strings.TrimSpace(c.From) == "" {
		return fmt.Errorf("%w: from required", ErrInvalidSMTPConfig)
	}
	return nil
}

// SMTPMailer sends emails via an SMTP server using wneessen/go-mail.
type SMTPMailer struct {
	cfg SMTPConfig
}

// NewSMTPMailer validates cfg and returns a ready-to-use mailer.
func NewSMTPMailer(cfg SMTPConfig) (*SMTPMailer, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &SMTPMailer{cfg: cfg}, nil
}

func (m *SMTPMailer) buildClient() (*gomail.Client, error) {
	opts := []gomail.Option{gomail.WithPort(m.cfg.Port)}
	switch strings.ToLower(strings.TrimSpace(m.cfg.Security)) {
	case "ssl":
		opts = append(opts, gomail.WithSSL())
	case "plain", "none":
		opts = append(opts, gomail.WithTLSPolicy(gomail.NoTLS))
	default: // "starttls" or unset
		opts = append(opts, gomail.WithTLSPolicy(gomail.TLSMandatory))
	}
	if strings.TrimSpace(m.cfg.Username) != "" {
		opts = append(opts,
			gomail.WithSMTPAuth(gomail.SMTPAuthLogin),
			gomail.WithUsername(m.cfg.Username),
			gomail.WithPassword(m.cfg.Password),
		)
	}
	return gomail.NewClient(m.cfg.Host, opts...)
}

func (m *SMTPMailer) Send(ctx context.Context, msg domainmail.Message) error {
	if strings.TrimSpace(msg.To) == "" {
		return errors.New("mail: empty recipient")
	}
	client, err := m.buildClient()
	if err != nil {
		return fmt.Errorf("mail: build client: %w", err)
	}
	gm := gomail.NewMsg()
	if strings.TrimSpace(m.cfg.FromName) != "" {
		if err := gm.FromFormat(m.cfg.FromName, m.cfg.From); err != nil {
			return fmt.Errorf("mail: from: %w", err)
		}
	} else {
		if err := gm.From(m.cfg.From); err != nil {
			return fmt.Errorf("mail: from: %w", err)
		}
	}
	if err := gm.To(msg.To); err != nil {
		return fmt.Errorf("mail: to: %w", err)
	}
	gm.Subject(msg.Subject)
	if strings.TrimSpace(msg.HTMLBody) != "" {
		gm.SetBodyString(gomail.TypeTextPlain, msg.TextBody)
		gm.AddAlternativeString(gomail.TypeTextHTML, msg.HTMLBody)
	} else {
		gm.SetBodyString(gomail.TypeTextPlain, msg.TextBody)
	}
	return client.DialAndSendWithContext(ctx, gm)
}
