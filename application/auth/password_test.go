package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"golang.org/x/crypto/bcrypt"

	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/infrastructure/persistence/gormdb"
	"github.com/lpxxn/blink/internal/testutil"
)

// fakeCodes is an in-memory CodeSendVerifier for tests. It records sends and
// uses a fixed code "123456" so tests can exercise Verify deterministically.
type fakeCodes struct {
	sends []fakeSend
	valid map[string]string // key = purpose|email -> code
}

type fakeSend struct {
	purpose string
	email   string
	actor   string
}

func newFakeCodes() *fakeCodes { return &fakeCodes{valid: map[string]string{}} }

func (f *fakeCodes) Send(_ context.Context, purpose, email, actor string) error {
	f.sends = append(f.sends, fakeSend{purpose, email, actor})
	f.valid[purpose+"|"+email] = "123456"
	return nil
}

func (f *fakeCodes) Verify(_ context.Context, purpose, email, code string) error {
	want, ok := f.valid[purpose+"|"+email]
	if !ok || want != code {
		return errors.New("invalid code")
	}
	delete(f.valid, purpose+"|"+email)
	return nil
}

func seedPasswordUser(t *testing.T, u *gormdb.UserRepository, node *snowflake.Node, email, password string) int64 {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	id := node.Generate().Int64()
	if err := u.Create(context.Background(), &domainuser.User{
		SnowflakeID:  id,
		Email:        email,
		Name:         email,
		PasswordHash: string(hash),
		Status:       domainuser.StatusActive,
		Role:         "user",
	}); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestPassword_ChangeSuccess_InvalidatesSessions(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	sess := &stubSessionStore{}
	codes := newFakeCodes()
	uid := seedPasswordUser(t, u, node, "a@x.com", "password12")
	tok, err := sess.Create(context.Background(), uid, time.Hour, "", "")
	if err != nil {
		t.Fatal(err)
	}
	svc := &PasswordService{Users: u, Sessions: sess, Codes: codes}
	if err := svc.SendChangeCode(context.Background(), uid); err != nil {
		t.Fatalf("send: %v", err)
	}
	if err := svc.ChangePassword(context.Background(), uid, "123456", "newpassword12"); err != nil {
		t.Fatalf("change: %v", err)
	}
	if _, err := sess.Get(context.Background(), tok); err == nil {
		t.Fatal("expected session revoked")
	}
	got, err := u.FindByEmail(context.Background(), "a@x.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte("newpassword12")); err != nil {
		t.Fatalf("new password not stored: %v", err)
	}
}

func TestPassword_ChangeWrongCode(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	codes := newFakeCodes()
	uid := seedPasswordUser(t, u, node, "a@x.com", "password12")
	svc := &PasswordService{Users: u, Codes: codes}
	if err := svc.SendChangeCode(context.Background(), uid); err != nil {
		t.Fatal(err)
	}
	if err := svc.ChangePassword(context.Background(), uid, "000000", "newpassword12"); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}
}

func TestPassword_ChangeWeakPassword(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	codes := newFakeCodes()
	uid := seedPasswordUser(t, u, node, "a@x.com", "password12")
	svc := &PasswordService{Users: u, Codes: codes}
	_ = svc.SendChangeCode(context.Background(), uid)
	if err := svc.ChangePassword(context.Background(), uid, "123456", "short"); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestPassword_ResetFlow(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	sess := &stubSessionStore{}
	codes := newFakeCodes()
	uid := seedPasswordUser(t, u, node, "a@x.com", "password12")
	tok, _ := sess.Create(context.Background(), uid, time.Hour, "", "")

	svc := &PasswordService{Users: u, Sessions: sess, Codes: codes}
	if err := svc.SendResetCode(context.Background(), "a@x.com", "ip"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ResetPassword(context.Background(), "a@x.com", "123456", "newpassword12"); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if _, err := sess.Get(context.Background(), tok); err == nil {
		t.Fatal("expected session revoked after reset")
	}
}

func TestPassword_ResetUnknownEmailDoesNotSend(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	u := &gormdb.UserRepository{DB: db}
	codes := newFakeCodes()
	svc := &PasswordService{Users: u, Codes: codes}
	if err := svc.SendResetCode(context.Background(), "ghost@x.com", "ip"); err != nil {
		t.Fatalf("should not error, got %v", err)
	}
	if len(codes.sends) != 0 {
		t.Fatal("should not have sent any mail for unknown email")
	}
}

func TestPassword_ResetUnknownEmailReturnsInvalidCode(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	u := &gormdb.UserRepository{DB: db}
	codes := newFakeCodes()
	svc := &PasswordService{Users: u, Codes: codes}
	if err := svc.ResetPassword(context.Background(), "ghost@x.com", "000000", "newpassword12"); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}
}
