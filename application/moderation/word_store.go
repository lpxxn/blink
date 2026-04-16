package moderation

import (
	"context"

	domainsensitiveword "github.com/lpxxn/blink/domain/sensitiveword"
)

// WordListStore loads enabled sensitive words from the database into the process-wide snapshot.
type WordListStore struct {
	Repo domainsensitiveword.Repository
}

// Reload refreshes SensitiveWords from ListEnabledWords.
func (s *WordListStore) Reload(ctx context.Context) error {
	if s == nil || s.Repo == nil {
		SetWordsSnapshot(nil)
		return nil
	}
	words, err := s.Repo.ListEnabledWords(ctx)
	if err != nil {
		return err
	}
	SetWordsSnapshot(words)
	return nil
}
