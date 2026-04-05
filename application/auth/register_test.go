package auth

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"

	"github.com/lpxxn/blink/infrastructure/persistence/gormdb"
	"github.com/lpxxn/blink/internal/testutil"
	domainsession "github.com/lpxxn/blink/domain/session"
)

// stubSessionStore avoids importing redisstore (would cycle via application/idp).
type stubSessionStore struct {
	tokens map[string]int64
}

func (s *stubSessionStore) Create(_ context.Context, userID int64, _ time.Duration, _, _ string) (string, error) {
	tok := "stub-" + strconv.FormatInt(userID, 10)
	if s.tokens == nil {
		s.tokens = make(map[string]int64)
	}
	s.tokens[tok] = userID
	return tok, nil
}

func (s *stubSessionStore) Get(_ context.Context, token string) (*domainsession.Session, error) {
	uid, ok := s.tokens[token]
	if !ok {
		return nil, domainsession.ErrNotFound
	}
	return &domainsession.Session{ID: token, UserID: uid, ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (s *stubSessionStore) Delete(_ context.Context, token string) error {
	delete(s.tokens, token)
	return nil
}

func TestRegisterWithPassword_CreatesBuiltinIdentity(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, err := snowflake.NewNode(1)
	if err != nil {
		t.Fatal(err)
	}
	u := &gormdb.UserRepository{DB: db}
	o := &gormdb.OAuthRepository{DB: db}
	svc := &RegisterService{
		Users:      u,
		Identities: o,
		Node:       node,
		Tx:         &gormdb.TxRunner{DB: db},
	}
	ctx := context.Background()
	id, err := svc.RegisterWithPassword(ctx, "u@example.com", "password12", "U")
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("expected user id")
	}
	sub := FormatBuiltinSubject(id)
	oid, err := o.FindByProviderSubject(ctx, "builtin", sub)
	if err != nil {
		t.Fatal(err)
	}
	if oid.UserID != id {
		t.Fatalf("user id mismatch")
	}
}

func TestRegisterWithPassword_DuplicateEmail(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	o := &gormdb.OAuthRepository{DB: db}
	svc := &RegisterService{
		Users: u, Identities: o, Node: node,
		Tx: &gormdb.TxRunner{DB: db},
	}
	ctx := context.Background()
	if _, err := svc.RegisterWithPassword(ctx, "u@example.com", "password12", "U"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.RegisterWithPassword(ctx, "u@example.com", "password12", "U")
	if err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegisterWithSession_ReturnsToken(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	o := &gormdb.OAuthRepository{DB: db}
	sess := &stubSessionStore{}
	svc := &RegisterService{
		Users:      u,
		Identities: o,
		Node:       node,
		Tx:         &gormdb.TxRunner{DB: db},
		Sessions:   sess,
		SessionTTL: time.Hour,
	}
	ctx := context.Background()
	uid, tok, err := svc.RegisterWithSession(ctx, "sess@example.com", "password12", "S", "127.0.0.1", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("expected session token")
	}
	got, err := sess.Get(ctx, tok)
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != uid {
		t.Fatalf("session user_id want %d got %d", uid, got.UserID)
	}
}

func TestRegisterWithSession_ErrSessionNotConfigured(t *testing.T) {
	ctx := context.Background()
	_, _, err := (&RegisterService{Node: mustNode(t)}).RegisterWithSession(ctx, "a@b.com", "password12", "A", "", "")
	if !errors.Is(err, ErrSessionNotConfigured) {
		t.Fatalf("expected ErrSessionNotConfigured, got %v", err)
	}
}

func mustNode(t *testing.T) *snowflake.Node {
	t.Helper()
	n, err := snowflake.NewNode(1)
	if err != nil {
		t.Fatal(err)
	}
	return n
}
