package moderation

import "errors"

// ErrSensitiveContent is returned when reply text must not be stored (e.g. sensitive words).
var ErrSensitiveContent = errors.New("moderation: sensitive content")
