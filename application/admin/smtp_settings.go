package admin

import (
	"context"
	"errors"
	"strconv"
	"strings"

	domainmail "github.com/lpxxn/blink/domain/mail"
	domainsettings "github.com/lpxxn/blink/domain/settings"
	"github.com/lpxxn/blink/infrastructure/mail"
)

// SMTPSettings is a thin facade for the admin UI to read / write SMTP
// settings via the existing app_settings table. The stored password is never
// returned; callers may only *set* it (or leave it blank to keep the previous
// value).
type SMTPSettings struct {
	Settings domainsettings.Repository
	// Mailer is optional. When set, Set() calls Invalidate() on it so the
	// next Send() rebuilds the SMTP client with the new config.
	Mailer *mail.ConfigurableMailer
}

// SMTPConfigView is what the admin UI sees when reading settings. Password is
// never returned; HasPassword indicates whether one is stored.
type SMTPConfigView struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	HasPassword bool   `json:"has_password"`
	From        string `json:"from"`
	FromName    string `json:"from_name"`
	Security    string `json:"security"`
}

// SMTPConfigUpdate is what the admin UI sends. An empty Password means "keep
// the existing one"; to clear, the UI sends "__clear__" (not implemented yet).
type SMTPConfigUpdate struct {
	Enabled  *bool   `json:"enabled"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	From     *string `json:"from"`
	FromName *string `json:"from_name"`
	Security *string `json:"security"`
}

func (s *SMTPSettings) Get(ctx context.Context) (SMTPConfigView, error) {
	if s.Settings == nil {
		return SMTPConfigView{}, ErrSettingsNotConfigured
	}
	get := func(key string) (string, error) {
		v, err := s.Settings.GetString(ctx, key)
		if errors.Is(err, domainsettings.ErrNotFound) {
			return "", nil
		}
		return strings.TrimSpace(v), err
	}
	enabledStr, err := get(mail.SettingEnabled)
	if err != nil {
		return SMTPConfigView{}, err
	}
	host, err := get(mail.SettingHost)
	if err != nil {
		return SMTPConfigView{}, err
	}
	portStr, err := get(mail.SettingPort)
	if err != nil {
		return SMTPConfigView{}, err
	}
	user, err := get(mail.SettingUsername)
	if err != nil {
		return SMTPConfigView{}, err
	}
	pw, err := get(mail.SettingPassword)
	if err != nil {
		return SMTPConfigView{}, err
	}
	from, err := get(mail.SettingFrom)
	if err != nil {
		return SMTPConfigView{}, err
	}
	fromName, err := get(mail.SettingFromName)
	if err != nil {
		return SMTPConfigView{}, err
	}
	security, err := get(mail.SettingSecurity)
	if err != nil {
		return SMTPConfigView{}, err
	}
	port, _ := strconv.Atoi(portStr)
	return SMTPConfigView{
		Enabled:     strings.EqualFold(enabledStr, "true"),
		Host:        host,
		Port:        port,
		Username:    user,
		HasPassword: pw != "",
		From:        from,
		FromName:    fromName,
		Security:    security,
	}, nil
}

// Set applies the update. All non-nil fields are written. When Password is nil
// or the empty string, the previously stored value is kept. smtp.version is
// incremented so running ConfigurableMailer instances reload quickly.
func (s *SMTPSettings) Set(ctx context.Context, u SMTPConfigUpdate) error {
	if s.Settings == nil {
		return ErrSettingsNotConfigured
	}
	if u.Enabled != nil {
		v := "false"
		if *u.Enabled {
			v = "true"
		}
		if err := s.Settings.SetString(ctx, mail.SettingEnabled, v); err != nil {
			return err
		}
	}
	if u.Host != nil {
		if err := s.Settings.SetString(ctx, mail.SettingHost, strings.TrimSpace(*u.Host)); err != nil {
			return err
		}
	}
	if u.Port != nil {
		if *u.Port <= 0 || *u.Port > 65535 {
			return ErrInvalidSetting
		}
		if err := s.Settings.SetString(ctx, mail.SettingPort, strconv.Itoa(*u.Port)); err != nil {
			return err
		}
	}
	if u.Username != nil {
		if err := s.Settings.SetString(ctx, mail.SettingUsername, strings.TrimSpace(*u.Username)); err != nil {
			return err
		}
	}
	if u.Password != nil && *u.Password != "" {
		if err := s.Settings.SetString(ctx, mail.SettingPassword, *u.Password); err != nil {
			return err
		}
	}
	if u.From != nil {
		if err := s.Settings.SetString(ctx, mail.SettingFrom, strings.TrimSpace(*u.From)); err != nil {
			return err
		}
	}
	if u.FromName != nil {
		if err := s.Settings.SetString(ctx, mail.SettingFromName, strings.TrimSpace(*u.FromName)); err != nil {
			return err
		}
	}
	if u.Security != nil {
		sec := strings.ToLower(strings.TrimSpace(*u.Security))
		switch sec {
		case "", "starttls", "ssl", "plain", "none":
		default:
			return ErrInvalidSetting
		}
		if err := s.Settings.SetString(ctx, mail.SettingSecurity, sec); err != nil {
			return err
		}
	}
	if err := s.bumpVersion(ctx); err != nil {
		return err
	}
	if s.Mailer != nil {
		s.Mailer.Invalidate()
	}
	return nil
}

// Test sends a test email to `to` using the currently configured SMTP. It
// bypasses the ConfigurableMailer cache so operators see a fresh attempt.
func (s *SMTPSettings) Test(ctx context.Context, to string) error {
	to = strings.TrimSpace(to)
	if to == "" || !strings.Contains(to, "@") {
		return errors.New("admin: invalid recipient")
	}
	if s.Mailer == nil {
		return errors.New("admin: mailer not configured")
	}
	s.Mailer.Invalidate()
	return s.Mailer.Send(ctx, domainmail.Message{
		To:       to,
		Subject:  "Blink SMTP test",
		TextBody: "This is a test email from Blink admin. If you received it, SMTP is configured correctly.",
		HTMLBody: `<p>This is a test email from <b>Blink</b> admin. If you received it, SMTP is configured correctly.</p>`,
	})
}

func (s *SMTPSettings) bumpVersion(ctx context.Context) error {
	cur, err := s.Settings.GetString(ctx, mail.SettingVersion)
	if err != nil && !errors.Is(err, domainsettings.ErrNotFound) {
		return err
	}
	v := 0
	if cur != "" {
		v, _ = strconv.Atoi(strings.TrimSpace(cur))
	}
	v++
	return s.Settings.SetString(ctx, mail.SettingVersion, strconv.Itoa(v))
}
