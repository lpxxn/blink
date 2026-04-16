package eventing

import "context"

// SensitiveWordsPublisher notifies other instances to reload the in-memory word list.
type SensitiveWordsPublisher interface {
	PublishSensitiveWordsChanged(ctx context.Context) error
}
