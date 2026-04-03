package redisstore

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	domainoauth "github.com/lpxxn/blink/domain/oauth"
)

func TestOAuthStateStore_ConsumeTwiceFails(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	s := &OAuthStateStore{Client: rdb}
	ctx := context.Background()
	if err := s.Save(ctx, "st1", domainoauth.RedirectState{Provider: "p", NextURL: "/"}, time.Minute); err != nil {
		t.Fatal(err)
	}
	p, err := s.Consume(ctx, "st1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Provider != "p" {
		t.Fatalf("provider: %q", p.Provider)
	}
	_, err = s.Consume(ctx, "st1")
	if err != domainoauth.ErrInvalidState {
		t.Fatalf("expected ErrInvalidState, got %v", err)
	}
}
