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

func createUser(t *testing.T, u *gormdb.UserRepository, node *snowflake.Node, email, password string, status int) int64 {
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
		Status:       status,
		Role:         "user",
	}); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestLogin_HappyPath(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	sess := &stubSessionStore{}
	uid := createUser(t, u, node, "a@x.com", "password12", domainuser.StatusActive)
	svc := &LoginService{Users: u, Sessions: sess, SessionTTL: time.Hour}
	got, tok, err := svc.LoginWithPassword(context.Background(), "A@X.com", "password12", "", "")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if got != uid {
		t.Fatalf("uid: got %d want %d", got, uid)
	}
	s, err := sess.Get(context.Background(), tok)
	if err != nil || s.UserID != uid {
		t.Fatalf("session lookup failed: %v %+v", err, s)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	sess := &stubSessionStore{}
	createUser(t, u, node, "a@x.com", "password12", domainuser.StatusActive)
	svc := &LoginService{Users: u, Sessions: sess, SessionTTL: time.Hour}
	_, _, err := svc.LoginWithPassword(context.Background(), "a@x.com", "wrong-password", "", "")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	u := &gormdb.UserRepository{DB: db}
	sess := &stubSessionStore{}
	svc := &LoginService{Users: u, Sessions: sess, SessionTTL: time.Hour}
	_, _, err := svc.LoginWithPassword(context.Background(), "nope@x.com", "password12", "", "")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	u := &gormdb.UserRepository{DB: db}
	sess := &stubSessionStore{}
	createUser(t, u, node, "a@x.com", "password12", domainuser.StatusBanned)
	svc := &LoginService{Users: u, Sessions: sess, SessionTTL: time.Hour}
	_, _, err := svc.LoginWithPassword(context.Background(), "a@x.com", "password12", "", "")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for banned, got %v", err)
	}
}
