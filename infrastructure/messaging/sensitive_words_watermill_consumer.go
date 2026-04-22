package messaging

import (
	"context"
	"errors"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// RunSensitiveWordsWatermillRouter consumes TopicSensitiveWords and triggers a reload callback.
func RunSensitiveWordsWatermillRouter(ctx context.Context, sub message.Subscriber, reload func(context.Context) error, wmLog watermill.LoggerAdapter) (*message.Router, error) {
	if reload == nil {
		return nil, errors.New("messaging: sensitive words reload func is nil")
	}
	router, err := message.NewRouter(message.RouterConfig{}, wmLog)
	if err != nil {
		return nil, err
	}
	router.AddConsumerHandler(
		"blink_sensitive_words_reload",
		TopicSensitiveWords,
		sub,
		func(_ *message.Message) error {
			if err := reload(context.Background()); err != nil {
				log.Printf("sensitive words reload: %v", err)
				return err
			}
			return nil
		},
	)
	go func() {
		if err := router.Run(ctx); err != nil {
			log.Printf("watermill router stopped: %v", err)
		}
	}()
	<-router.Running()
	return router, nil
}
