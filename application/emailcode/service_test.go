package emailcode

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	domainmail "github.com/lpxxn/blink/domain/mail"
	"github.com/lpxxn/blink/infrastructure/cache/redisstore"
)

type recordingMailer struct {
	mu       sync.Mutex
	messages []domainmail.Message
	failWith error
}

func (r *recordingMailer) Send(_ context.Context, msg domainmail.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failWith != nil {
		return r.failWith
	}
	r.messages = append(r.messages, msg)
	return nil
}

func (r *recordingMailer) last() (domainmail.Message, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.messages) == 0 {
		return domainmail.Message{}, false
	}
	return r.messages[len(r.messages)-1], true
}

func (r *recordingMailer) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.messages)
}

func newServiceT(t *testing.T) (*Service, *miniredis.Miniredis, *recordingMailer) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mr.Close() })
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	store := &redisstore.EmailCodeStore{Client: rdb}
	m := &recordingMailer{}
	svc := &Service{Store: store, Mailer: m}
	return svc, mr, m
}

// extractCode pulls the 6-digit code out of the mailer's rendered text body.
func extractCode(t *testing.T, body string) string {
	t.Helper()
	idx := strings.Index(body, "验证码：")
	if idx < 0 {
		t.Fatalf("no code marker in body: %q", body)
	}
	s := body[idx+len("验证码："):]
	if len(s) < 6 {
		t.Fatalf("body too short after marker: %q", s)
	}
	return s[:6]
}

func TestService_SendAndVerify(t *testing.T) {
	svc, _, mail := newServiceT(t)
	ctx := context.Background()
	if err := svc.Send(ctx, PurposeRegister, "Alice@Example.com", "127.0.0.1"); err != nil {
		t.Fatalf("send: %v", err)
	}
	msg, ok := mail.last()
	if !ok {
		t.Fatal("expected mail sent")
	}
	if msg.To != "alice@example.com" {
		t.Fatalf("normalized to: %q", msg.To)
	}
	code := extractCode(t, msg.TextBody)
	if err := svc.Verify(ctx, PurposeRegister, "alice@example.com", code); err != nil {
		t.Fatalf("verify: %v", err)
	}
	// Second verify should fail (one-shot).
	if err := svc.Verify(ctx, PurposeRegister, "alice@example.com", code); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode on reuse, got %v", err)
	}
}

func TestService_PurposeIsolated(t *testing.T) {
	svc, _, mail := newServiceT(t)
	ctx := context.Background()
	if err := svc.Send(ctx, PurposeRegister, "a@x.com", ""); err != nil {
		t.Fatal(err)
	}
	msg, _ := mail.last()
	code := extractCode(t, msg.TextBody)
	if err := svc.Verify(ctx, PurposeChangePassword, "a@x.com", code); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("wrong purpose should not verify: %v", err)
	}
}

func TestService_CoolDown(t *testing.T) {
	svc, mr, _ := newServiceT(t)
	ctx := context.Background()
	if err := svc.Send(ctx, PurposeRegister, "a@x.com", ""); err != nil {
		t.Fatal(err)
	}
	if err := svc.Send(ctx, PurposeRegister, "a@x.com", ""); !errors.Is(err, ErrCoolingDown) {
		t.Fatalf("expected cooling down, got %v", err)
	}
	mr.FastForward(61 * time.Second)
	if err := svc.Send(ctx, PurposeRegister, "a@x.com", ""); err != nil {
		t.Fatalf("after cooldown expected ok, got %v", err)
	}
}

func TestService_HourlyEmailLimit(t *testing.T) {
	svc, mr, _ := newServiceT(t)
	ctx := context.Background()
	const email = "a@x.com"
	for i := 0; i < 5; i++ {
		if err := svc.Send(ctx, PurposeRegister, email, ""); err != nil {
			t.Fatalf("send #%d: %v", i, err)
		}
		mr.FastForward(61 * time.Second) // bypass cooldown between requests
	}
	if err := svc.Send(ctx, PurposeRegister, email, ""); !errors.Is(err, ErrTooMany) {
		t.Fatalf("expected ErrTooMany on 6th send, got %v", err)
	}
	mr.FastForward(time.Hour + time.Second)
	if err := svc.Send(ctx, PurposeRegister, email, ""); err != nil {
		t.Fatalf("after window reset expected ok, got %v", err)
	}
}

func TestService_FailLockout(t *testing.T) {
	svc, _, mail := newServiceT(t)
	ctx := context.Background()
	if err := svc.Send(ctx, PurposeRegister, "a@x.com", ""); err != nil {
		t.Fatal(err)
	}
	msg, _ := mail.last()
	realCode := extractCode(t, msg.TextBody)
	for i := 0; i < 5; i++ {
		if err := svc.Verify(ctx, PurposeRegister, "a@x.com", "000000"); !errors.Is(err, ErrInvalidCode) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	// After 5 failed attempts the code should be invalidated even with the real code.
	if err := svc.Verify(ctx, PurposeRegister, "a@x.com", realCode); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("code should be locked out, got %v", err)
	}
}

func TestService_MailerFailureDoesNotConsumeLimit(t *testing.T) {
	svc, _, mail := newServiceT(t)
	ctx := context.Background()
	mail.failWith = errors.New("smtp down")
	if err := svc.Send(ctx, PurposeRegister, "a@x.com", ""); err == nil {
		t.Fatal("expected mailer error")
	}
	mail.failWith = nil
	// After the mailer recovered, the same email should send without being
	// blocked by the cooldown/limit counters (we did not write them on failure).
	if err := svc.Send(ctx, PurposeRegister, "a@x.com", ""); err != nil {
		t.Fatalf("expected ok after mailer recovered, got %v", err)
	}
	if mail.count() != 1 {
		t.Fatalf("expected 1 successful send, got %d", mail.count())
	}
}

func TestService_UnknownPurposeAndBadEmail(t *testing.T) {
	svc, _, _ := newServiceT(t)
	ctx := context.Background()
	if err := svc.Send(ctx, "nope", "a@x.com", ""); !errors.Is(err, ErrUnknownPurpose) {
		t.Fatalf("purpose: %v", err)
	}
	if err := svc.Send(ctx, PurposeRegister, "bad-email", ""); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("email: %v", err)
	}
}
