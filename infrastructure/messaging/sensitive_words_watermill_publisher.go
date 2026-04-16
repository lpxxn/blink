package messaging

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	appeventing "github.com/lpxxn/blink/application/eventing"
)

// SensitiveWordsWatermillPublisher publishes reload signals on TopicSensitiveWords.
type SensitiveWordsWatermillPublisher struct {
	inner message.Publisher
}

func NewSensitiveWordsWatermillPublisher(pub message.Publisher) *SensitiveWordsWatermillPublisher {
	return &SensitiveWordsWatermillPublisher{inner: pub}
}

var _ appeventing.SensitiveWordsPublisher = (*SensitiveWordsWatermillPublisher)(nil)

func (p *SensitiveWordsWatermillPublisher) PublishSensitiveWordsChanged(ctx context.Context) error {
	if p == nil || p.inner == nil {
		return nil
	}
	payload, err := json.Marshal(map[string]string{"type": "sensitive_words_reload"})
	if err != nil {
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), payload)
	return p.inner.Publish(TopicSensitiveWords, msg)
}
