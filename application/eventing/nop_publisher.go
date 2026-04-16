package eventing

import "context"

// NopNotificationPublisher discards all events (tests or disabled messaging).
type NopNotificationPublisher struct{}

func (NopNotificationPublisher) PublishReplyToPost(context.Context, int64, int64, int64, string) error {
	return nil
}

func (NopNotificationPublisher) PublishReplyToComment(context.Context, int64, int64, int64, string) error {
	return nil
}

func (NopNotificationPublisher) PublishPostRemoved(context.Context, int64, int64, string) error {
	return nil
}

func (NopNotificationPublisher) PublishPostFlagged(context.Context, int64, int64, string) error {
	return nil
}

func (NopNotificationPublisher) PublishSensitiveHitForAdmins(context.Context, int64, int64, []string) error {
	return nil
}

func (NopNotificationPublisher) PublishAppealSubmitted(context.Context, int64, int64, string, string) error {
	return nil
}

func (NopNotificationPublisher) PublishAppealResolved(context.Context, int64, int64, bool, string) error {
	return nil
}

func (NopNotificationPublisher) PublishUserBanned(context.Context, int64) error {
	return nil
}

// NopSensitiveWordsPublisher discards sensitive-word reload signals (tests).
type NopSensitiveWordsPublisher struct{}

func (NopSensitiveWordsPublisher) PublishSensitiveWordsChanged(context.Context) error {
	return nil
}
