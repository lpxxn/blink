package mail

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	domainmail "github.com/lpxxn/blink/domain/mail"
	domainsettings "github.com/lpxxn/blink/domain/settings"
)

// Settings key constants. Keep in sync with docs/email-auth.md and admin UI.
const (
	SettingEnabled  = "smtp.enabled"
	SettingHost     = "smtp.host"
	SettingPort     = "smtp.port"
	SettingUsername = "smtp.username"
	SettingPassword = "smtp.password"
	SettingFrom     = "smtp.from"
	SettingFromName = "smtp.from_name"
	SettingSecurity = "smtp.security"
	SettingVersion  = "smtp.version"
)

// ConfigurableMailer reads SMTP settings from the domain settings repo on
// each send. It caches the constructed SMTPMailer until smtp.version changes
// (or the reloadInterval elapses) so UI updates take effect almost immediately
// without restarting the process.
//
// All configuration (including the SMTP password) lives in the app_settings
// table. Operators who need encryption-at-rest should rely on the storage
// layer (encrypted volume, DB-level encryption); this package does not add
// its own envelope encryption because that would require an extra secret
// bootstrap channel and the user explicitly asked for zero env config.
type ConfigurableMailer struct {
	Settings domainsettings.Repository
	// Fallback is used when SMTP is disabled / unconfigured or when building
	// the SMTP client fails.
	Fallback domainmail.Mailer
	// ReloadInterval is a safety net in case smtp.version tracking drifts.
	// Defaults to 30 seconds. At most one reload happens per interval even
	// if version is unchanged.
	ReloadInterval time.Duration

	mu         sync.Mutex
	cached     *SMTPMailer
	cachedVer  string
	cachedAt   time.Time
	lastErrLog time.Time
}

func (m *ConfigurableMailer) reloadInterval() time.Duration {
	if m.ReloadInterval > 0 {
		return m.ReloadInterval
	}
	return 30 * time.Second
}

// LoadConfig reads the current SMTP settings from the repo. The returned
// `enabled` is false when smtp.enabled != "true" or smtp.host is empty.
func (m *ConfigurableMailer) LoadConfig(ctx context.Context) (cfg SMTPConfig, enabled bool, version string, err error) {
	get := func(key string) (string, error) {
		v, e := m.Settings.GetString(ctx, key)
		if errors.Is(e, domainsettings.ErrNotFound) {
			return "", nil
		}
		return strings.TrimSpace(v), e
	}
	enabledStr, err := get(SettingEnabled)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	host, err := get(SettingHost)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	portStr, err := get(SettingPort)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	user, err := get(SettingUsername)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	pw, err := get(SettingPassword)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	from, err := get(SettingFrom)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	fromName, err := get(SettingFromName)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	security, err := get(SettingSecurity)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	version, err = get(SettingVersion)
	if err != nil {
		return SMTPConfig{}, false, "", err
	}
	port := 587
	if portStr != "" {
		if v, convErr := strconv.Atoi(portStr); convErr == nil {
			port = v
		}
	}
	cfg = SMTPConfig{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pw,
		From:     from,
		FromName: fromName,
		Security: security,
	}
	enabled = strings.EqualFold(enabledStr, "true") && host != ""
	return cfg, enabled, version, nil
}

// resolve returns the currently-configured SMTP mailer or (nil, false) when
// the caller should fall back. Uses a version/ttl cache.
func (m *ConfigurableMailer) resolve(ctx context.Context) (*SMTPMailer, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cached != nil && time.Since(m.cachedAt) < m.reloadInterval() {
		return m.cached, true
	}
	cfg, enabled, ver, err := m.LoadConfig(ctx)
	if err != nil {
		m.logRateLimited("mail: load smtp config: %v", err)
		return nil, false
	}
	if !enabled {
		m.cached = nil
		m.cachedVer = ver
		m.cachedAt = time.Now()
		return nil, false
	}
	if m.cached != nil && ver == m.cachedVer {
		m.cachedAt = time.Now()
		return m.cached, true
	}
	mailer, err := NewSMTPMailer(cfg)
	if err != nil {
		m.logRateLimited("mail: build smtp mailer: %v", err)
		m.cached = nil
		m.cachedVer = ver
		m.cachedAt = time.Now()
		return nil, false
	}
	m.cached = mailer
	m.cachedVer = ver
	m.cachedAt = time.Now()
	return mailer, true
}

// Invalidate drops the cached SMTP client so the next Send rebuilds it. Used
// by the admin handler after a PUT to settings.
func (m *ConfigurableMailer) Invalidate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cached = nil
	m.cachedVer = ""
	m.cachedAt = time.Time{}
}

func (m *ConfigurableMailer) logRateLimited(format string, args ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if time.Since(m.lastErrLog) < 10*time.Second {
		return
	}
	m.lastErrLog = time.Now()
	log.Printf(format, args...)
}

// Send implements domain/mail.Mailer.
func (m *ConfigurableMailer) Send(ctx context.Context, msg domainmail.Message) error {
	if active, ok := m.resolve(ctx); ok {
		return active.Send(ctx, msg)
	}
	if m.Fallback == nil {
		return errors.New("mail: SMTP is not configured and no fallback mailer is set")
	}
	return m.Fallback.Send(ctx, msg)
}
