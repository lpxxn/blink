package moderation

import "errors"

// ErrSensitiveContent is returned when reply text must not be stored (e.g. sensitive words).
var ErrSensitiveContent = errors.New("moderation: sensitive content")

// SensitiveContentError includes which configured words matched. errors.Is(err, ErrSensitiveContent) is true.
type SensitiveContentError struct {
	Hits []string
}

func (e *SensitiveContentError) Error() string {
	return ErrSensitiveContent.Error()
}

func (e *SensitiveContentError) Unwrap() error {
	return ErrSensitiveContent
}

// ErrSensitiveWithHits returns an error that unwraps to ErrSensitiveContent. hits must be non-empty.
func ErrSensitiveWithHits(hits []string) error {
	if len(hits) == 0 {
		return ErrSensitiveContent
	}
	cp := append([]string(nil), hits...)
	return &SensitiveContentError{Hits: cp}
}
