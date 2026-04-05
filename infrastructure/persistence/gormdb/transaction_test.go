package gormdb

import (
	"context"
	"errors"
	"testing"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/internal/testutil"
	"gorm.io/gorm"
)

// TestWithTransaction_Commit shows the usual pattern: inside the callback, build repos
// with {DB: tx} so user + oauth inserts are atomic and visible after commit.
func TestWithTransaction_Commit(t *testing.T) {
	gdb := testutil.OpenSQLiteMemory(t)
	ctx := context.Background()

	err := WithTransaction(gdb, func(tx *gorm.DB) error {
		users := &UserRepository{DB: tx}
		oauth := &OAuthRepository{DB: tx}

		if err := users.Create(ctx, &domainuser.User{
			SnowflakeID:  77001,
			Email:        "tx-ok@example.com",
			Name:         "Tx OK",
			PasswordHash: "h",
			Status:       domainuser.StatusActive,
			Role:         "user",
		}); err != nil {
			return err
		}
		return oauth.Create(ctx, &domainoauth.Identity{
			SnowflakeID:     77002,
			Provider:        "test",
			ProviderSubject: "sub-tx-ok",
			UserID:          77001,
		})
	})
	if err != nil {
		t.Fatal(err)
	}

	// After commit, use the root gdb (or a new Session) to read committed data.
	outUser := &UserRepository{DB: gdb}
	outOAuth := &OAuthRepository{DB: gdb}
	u, err := outUser.FindByEmail(ctx, "tx-ok@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.SnowflakeID != 77001 {
		t.Fatalf("user: %+v", u)
	}
	id, err := outOAuth.FindByProviderSubject(ctx, "test", "sub-tx-ok")
	if err != nil {
		t.Fatal(err)
	}
	if id.UserID != 77001 {
		t.Fatalf("identity: %+v", id)
	}
}

// TestWithTransaction_RollbackOnError shows that a failing fn rolls back all writes
// done on tx-backed repositories.
func TestWithTransaction_RollbackOnError(t *testing.T) {
	gdb := testutil.OpenSQLiteMemory(t)
	ctx := context.Background()

	err := WithTransaction(gdb, func(tx *gorm.DB) error {
		users := &UserRepository{DB: tx}
		if err := users.Create(ctx, &domainuser.User{
			SnowflakeID:  78001,
			Email:        "tx-fail@example.com",
			Name:         "Tx Fail",
			PasswordHash: "h",
			Status:       domainuser.StatusActive,
			Role:         "user",
		}); err != nil {
			return err
		}
		return errors.New("simulate business failure")
	})
	if err == nil {
		t.Fatal("expected error from transaction callback")
	}

	out := &UserRepository{DB: gdb}
	_, err = out.FindByEmail(ctx, "tx-fail@example.com")
	if !errors.Is(err, domainuser.ErrNotFound) {
		t.Fatalf("expected user absent after rollback, err=%v", err)
	}
}

// TestWithTransaction_NestedUsesSameTx documents that you can pass tx into helpers
// as *gorm.DB; do not mix r.DB and tx in one transaction.
func TestWithTransaction_HelperReceivesTx(t *testing.T) {
	gdb := testutil.OpenSQLiteMemory(t)
	ctx := context.Background()

	err := WithTransaction(gdb, func(tx *gorm.DB) error {
		return seedUserAndOAuth(ctx, tx, 79001, 79002, "helper@example.com", "h-sub")
	})
	if err != nil {
		t.Fatal(err)
	}

	u, _ := (&UserRepository{DB: gdb}).FindByEmail(ctx, "helper@example.com")
	if u == nil || u.SnowflakeID != 79001 {
		t.Fatalf("user after tx: %+v", u)
	}
}

func seedUserAndOAuth(ctx context.Context, db *gorm.DB, uid, oid int64, email, sub string) error {
	users := &UserRepository{DB: db}
	oauth := &OAuthRepository{DB: db}
	if err := users.Create(ctx, &domainuser.User{
		SnowflakeID:  uid, Email: email, Name: "H", PasswordHash: "x",
		Status: domainuser.StatusActive, Role: "user",
	}); err != nil {
		return err
	}
	return oauth.Create(ctx, &domainoauth.Identity{
		SnowflakeID: oid, Provider: "test", ProviderSubject: sub, UserID: uid,
	})
}
