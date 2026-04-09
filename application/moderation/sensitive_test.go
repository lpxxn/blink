package moderation

import "testing"

func TestFindSensitiveHits_emptyWords(t *testing.T) {
	if h := FindSensitiveHits("anything", nil); len(h) != 0 {
		t.Fatalf("got %v", h)
	}
}

func TestFindSensitiveHits_substring(t *testing.T) {
	words := []string{"敏感", "badword"}
	h := FindSensitiveHits("这里有敏感词", words)
	if len(h) != 1 || h[0] != "敏感" {
		t.Fatalf("got %v", h)
	}
}

func TestFindSensitiveHits_caseASCII(t *testing.T) {
	words := []string{"Spam"}
	h := FindSensitiveHits("this is SPAM here", words)
	if len(h) != 1 || h[0] != "Spam" {
		t.Fatalf("got %v", h)
	}
}

func TestPostModerationFromHits_pass(t *testing.T) {
	f, n := PostModerationFromHits(nil)
	if f != 0 || n != "" {
		t.Fatalf("flag=%d note=%q", f, n)
	}
}

func TestPostModerationFromHits_flag(t *testing.T) {
	f, n := PostModerationFromHits([]string{"a", "b"})
	if f != 1 || n == "" {
		t.Fatalf("flag=%d note=%q", f, n)
	}
}
