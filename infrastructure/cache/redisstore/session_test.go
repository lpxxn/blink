package redisstore

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	domainsession "github.com/lpxxn/blink/domain/session"
)

func TestSessionStore_CreateGetDelete(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	s := &SessionStore{Client: rdb}
	ctx := context.Background()
	tok, err := s.Create(ctx, 42, time.Minute, "127.0.0.1", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("expected token")
	}
	sess, err := s.Get(ctx, tok)
	if err != nil {
		t.Fatal(err)
	}
	if sess.UserID != 42 {
		t.Fatalf("user id: %d", sess.UserID)
	}
	if err := s.Delete(ctx, tok); err != nil {
		t.Fatal(err)
	}
	_, err = s.Get(ctx, tok)
	if err != domainsession.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
