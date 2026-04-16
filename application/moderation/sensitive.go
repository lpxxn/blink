package moderation

import (
	"strings"

	domainpost "github.com/lpxxn/blink/domain/post"
)

// SensitiveWords returns the current in-memory word list (reloaded from DB via WordListStore).
func SensitiveWords() []string {
	return loadWordsSnapshot()
}

// FindSensitiveHits returns which configured words appear as substrings in text.
// Matching is case-insensitive (strings.ToLower on both sides; Chinese unchanged).
func FindSensitiveHits(text string, words []string) []string {
	if text == "" || len(words) == 0 {
		return nil
	}
	lt := strings.ToLower(text)
	seen := make(map[string]struct{})
	var hits []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		lw := strings.ToLower(w)
		if !strings.Contains(lt, lw) {
			continue
		}
		if _, ok := seen[lw]; ok {
			continue
		}
		seen[lw] = struct{}{}
		hits = append(hits, w)
	}
	return hits
}

const maxModerationNoteLen = 2048

// ModerationNoteForSensitiveHits builds moderation_note when a post hits configured sensitive words.
func ModerationNoteForSensitiveHits(hits []string) string {
	if len(hits) == 0 {
		return ""
	}
	note := "有敏感词：" + strings.Join(hits, "、")
	if len(note) > maxModerationNoteLen {
		note = note[:maxModerationNoteLen]
	}
	return note
}

// PostModerationFromHits returns moderation_flag and moderation_note for posts.
// No hits → 审核通过 (ModerationNormal); otherwise flagged for review.
func PostModerationFromHits(hits []string) (flag int, note string) {
	if len(hits) == 0 {
		return domainpost.ModerationNormal, ""
	}
	return domainpost.ModerationFlagged, ModerationNoteForSensitiveHits(hits)
}

// ReplyContainsSensitive reports whether the reply body matches any configured word.
func ReplyContainsSensitive(body string, words []string) bool {
	return len(FindSensitiveHits(body, words)) > 0
}
