package httpapi

import (
	"testing"
	"time"

	domainfollow "github.com/lpxxn/blink/domain/follow"
)

func TestFollowListCursorRoundTrip(t *testing.T) {
	e := domainfollow.ListEntry{
		UserID:    9001,
		CreatedAt: time.Date(2026, 3, 1, 8, 30, 0, 123, time.UTC),
	}
	raw := formatFollowListCursor(e)
	got, err := parseFollowListCursor(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != e.UserID || !got.CreatedAt.Equal(e.CreatedAt) {
		t.Fatalf("round trip: got %+v want %+v", got, e)
	}
}

func TestParseFollowListCursor_Invalid(t *testing.T) {
	for _, raw := range []string{"", "abc", "1", "1,2,3"} {
		if _, err := parseFollowListCursor(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}
