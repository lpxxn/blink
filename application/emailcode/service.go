// Package emailcode issues and verifies one-time email verification codes.
// It is used by three flows: registration, logged-in change-password, and
// logged-out password reset. Codes are single-use and namespaced by purpose
// via a Redis-backed Store.
package emailcode

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	domainmail "github.com/lpxxn/blink/domain/mail"
)

const (
	PurposeRegister       = "register"
	PurposeChangePassword = "change_password"
	PurposeResetPassword  = "reset_password"

	defaultCodeTTL       = 10 * time.Minute
	defaultCoolDown      = 60 * time.Second
	defaultWindow        = time.Hour
	defaultMaxPerEmail   = 5
	defaultMaxPerActor   = 20
	defaultMaxFailVerify = 5
)

var (
	ErrInvalidEmail   = errors.New("emailcode: invalid email")
	ErrInvalidCode    = errors.New("emailcode: invalid or expired code")
	ErrCoolingDown    = errors.New("emailcode: please wait before requesting again")
	ErrTooMany        = errors.New("emailcode: too many requests, try later")
	ErrUnknownPurpose = errors.New("emailcode: unknown purpose")
)

// Store is the minimal storage contract required by Service. The production
// implementation lives in infrastructure/cache/redisstore.
type Store interface {
	IsCoolingDown(ctx context.Context, purpose, email string) (bool, error)
	HourlyEmailCount(ctx context.Context, purpose, email string) (int64, error)
	HourlyActorCount(ctx context.Context, purpose, actor string) (int64, error)
	PutCode(ctx context.Context, purpose, email, code, actor string, codeTTL, coolTTL, windowTTL time.Duration) error
	GetCode(ctx context.Context, purpose, email string) (string, error)
	IncrFail(ctx context.Context, purpose, email string, ttl time.Duration) (int64, error)
	ClearCode(ctx context.Context, purpose, email string) error
}

// Service owns the send/verify workflow. Zero values for the knobs fall back
// to production defaults.
type Service struct {
	Store  Store
	Mailer domainmail.Mailer

	CodeTTL       time.Duration
	CoolDown      time.Duration
	Window        time.Duration
	MaxPerEmail   int64
	MaxPerActor   int64
	MaxFailVerify int64

	// ProductName is the static fallback used in email subject/body when
	// ProductNameFn returns an empty string. Defaults to "Blink".
	ProductName string
	// ProductNameFn lets callers source the product name from app_settings so
	// admins can change it without restarting. Takes precedence over
	// ProductName when non-nil and returning a non-empty trimmed string.
	ProductNameFn func(context.Context) string

	// Now allows tests to override the clock (unused today but reserved).
	Now func() time.Time
}

func (s *Service) codeTTL() time.Duration {
	if s.CodeTTL > 0 {
		return s.CodeTTL
	}
	return defaultCodeTTL
}
func (s *Service) coolDown() time.Duration {
	if s.CoolDown > 0 {
		return s.CoolDown
	}
	return defaultCoolDown
}
func (s *Service) window() time.Duration {
	if s.Window > 0 {
		return s.Window
	}
	return defaultWindow
}
func (s *Service) maxPerEmail() int64 {
	if s.MaxPerEmail > 0 {
		return s.MaxPerEmail
	}
	return defaultMaxPerEmail
}
func (s *Service) maxPerActor() int64 {
	if s.MaxPerActor > 0 {
		return s.MaxPerActor
	}
	return defaultMaxPerActor
}
func (s *Service) maxFailVerify() int64 {
	if s.MaxFailVerify > 0 {
		return s.MaxFailVerify
	}
	return defaultMaxFailVerify
}

func (s *Service) productName(ctx context.Context) string {
	if s.ProductNameFn != nil {
		if v := strings.TrimSpace(s.ProductNameFn(ctx)); v != "" {
			return v
		}
	}
	if strings.TrimSpace(s.ProductName) != "" {
		return s.ProductName
	}
	return "Blink"
}

func validPurpose(p string) bool {
	switch p {
	case PurposeRegister, PurposeChangePassword, PurposeResetPassword:
		return true
	}
	return false
}

// NormalizeEmail lowercases and trims the email. Exposed for callers that
// need to use the same key shape (e.g. register.send_code reusing purpose).
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) validateEmail(email string) (string, error) {
	e := NormalizeEmail(email)
	if e == "" || !strings.Contains(e, "@") {
		return "", ErrInvalidEmail
	}
	return e, nil
}

// Send generates, stores, and emails a fresh code. actor is an opaque key used
// for per-IP (unauthenticated) or per-user (authenticated) rate limiting; may
// be empty to skip actor-level limiting.
func (s *Service) Send(ctx context.Context, purpose, email, actor string) error {
	if !validPurpose(purpose) {
		return ErrUnknownPurpose
	}
	e, err := s.validateEmail(email)
	if err != nil {
		return err
	}
	cooling, err := s.Store.IsCoolingDown(ctx, purpose, e)
	if err != nil {
		return err
	}
	if cooling {
		return ErrCoolingDown
	}
	n, err := s.Store.HourlyEmailCount(ctx, purpose, e)
	if err != nil {
		return err
	}
	if n >= s.maxPerEmail() {
		return ErrTooMany
	}
	if actor != "" {
		m, err := s.Store.HourlyActorCount(ctx, purpose, actor)
		if err != nil {
			return err
		}
		if m >= s.maxPerActor() {
			return ErrTooMany
		}
	}
	code, err := randomCode6()
	if err != nil {
		return err
	}
	subject, text, html := s.renderEmail(ctx, purpose, code)
	if err := s.Mailer.Send(ctx, domainmail.Message{To: e, Subject: subject, TextBody: text, HTMLBody: html}); err != nil {
		return err
	}
	return s.Store.PutCode(ctx, purpose, e, code, actor, s.codeTTL(), s.coolDown(), s.window())
}

// Verify checks the submitted code for (purpose,email). On success the code
// (and its failure counter) are deleted; on too many failures the code is
// invalidated even though the TTL has not elapsed.
func (s *Service) Verify(ctx context.Context, purpose, email, code string) error {
	if !validPurpose(purpose) {
		return ErrUnknownPurpose
	}
	e, err := s.validateEmail(email)
	if err != nil {
		return err
	}
	stored, err := s.Store.GetCode(ctx, purpose, e)
	if err != nil {
		return err
	}
	if stored == "" {
		return ErrInvalidCode
	}
	submitted := strings.TrimSpace(code)
	if subtle.ConstantTimeCompare([]byte(stored), []byte(submitted)) != 1 {
		n, ferr := s.Store.IncrFail(ctx, purpose, e, s.codeTTL())
		if ferr == nil && n >= s.maxFailVerify() {
			_ = s.Store.ClearCode(ctx, purpose, e)
		}
		return ErrInvalidCode
	}
	_ = s.Store.ClearCode(ctx, purpose, e)
	return nil
}

func (s *Service) renderEmail(ctx context.Context, purpose, code string) (subject, text, html string) {
	p := s.productName(ctx)
	ttl := int(s.codeTTL().Minutes())
	var scene string
	switch purpose {
	case PurposeRegister:
		scene = "完成注册"
	case PurposeChangePassword:
		scene = "修改登录密码"
	case PurposeResetPassword:
		scene = "重置登录密码"
	default:
		scene = "身份验证"
	}
	subject = fmt.Sprintf("[%s] 您的验证码：%s", p, code)
	text = fmt.Sprintf(
		"您正在 %s 执行%s操作。\n\n验证码：%s\n\n验证码 %d 分钟内有效，请勿泄露给他人。\n如果这不是您本人的操作，请忽略本邮件。",
		p, scene, code, ttl,
	)
	html = fmt.Sprintf(
		`<div style="font-family:system-ui,-apple-system,Segoe UI,Helvetica,Arial,sans-serif;font-size:14px;line-height:1.6;color:#222">
<p>您正在 <b>%s</b> 执行<b>%s</b>操作。</p>
<p>验证码：<span style="font-size:22px;letter-spacing:4px;font-weight:700;color:#111">%s</span></p>
<p>验证码 %d 分钟内有效，请勿泄露给他人。如果这不是您本人的操作，请忽略本邮件。</p>
</div>`, p, scene, code, ttl)
	return
}

// randomCode6 returns a uniformly random 6-digit string (000000..999999).
func randomCode6() (string, error) {
	max := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
