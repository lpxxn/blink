package moderation

import (
	"errors"
	"testing"
)

func TestErrSensitiveWithHits_Unwrap(t *testing.T) {
	err := ErrSensitiveWithHits([]string{"a", "b"})
	if !errors.Is(err, ErrSensitiveContent) {
		t.Fatalf("errors.Is: %v", err)
	}
	var sce *SensitiveContentError
	if !errors.As(err, &sce) {
		t.Fatal("expected As SensitiveContentError")
	}
	if len(sce.Hits) != 2 || sce.Hits[0] != "a" || sce.Hits[1] != "b" {
		t.Fatalf("hits=%v", sce.Hits)
	}
}
