package moderation

import (
	"sync/atomic"
)

var wordsSnapshot atomic.Value // []string

func init() {
	wordsSnapshot.Store([]string(nil))
}

// SetWordsSnapshot replaces the in-memory word list used by SensitiveWords (tests, reload).
func SetWordsSnapshot(words []string) {
	if len(words) == 0 {
		wordsSnapshot.Store([]string(nil))
		return
	}
	cp := append([]string(nil), words...)
	wordsSnapshot.Store(cp)
}

func loadWordsSnapshot() []string {
	v := wordsSnapshot.Load()
	if v == nil {
		return nil
	}
	return v.([]string)
}
