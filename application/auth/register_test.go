package auth

import (
	"context"
	"testing"

	"github.com/bwmarrin/snowflake"

	"github.com/lpxxn/blink/infrastructure/persistence/gormdb"
	"github.com/lpxxn/blink/internal/testutil"
)

func TestRegisterWithPassword_CreatesBuiltinIdentity(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, err := snowflake.NewNode(1)
	if err != nil {
		t.Fatal(err)
	}
	svc := &RegisterService{
		Users:      &gormdb.UserRepository{DB: db},
		Identities: &gormdb.OAuthRepository{DB: db},
		Node:       node,
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
	o, err := (&gormdb.OAuthRepository{DB: db}).FindByProviderSubject(ctx, "builtin", sub)
	if err != nil {
		t.Fatal(err)
	}
	if o.UserID != id {
		t.Fatalf("user id mismatch")
	}
}

func TestRegisterWithPassword_DuplicateEmail(t *testing.T) {
	db := testutil.OpenSQLiteMemory(t)
	node, _ := snowflake.NewNode(1)
	svc := &RegisterService{
		Users:      &gormdb.UserRepository{DB: db},
		Identities: &gormdb.OAuthRepository{DB: db},
		Node:       node,
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
