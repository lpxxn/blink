package httpapi

import (
	"errors"

	appmoderation "github.com/lpxxn/blink/application/moderation"
	"github.com/gin-gonic/gin"
)

// SensitiveContentPayload builds a 400 JSON body for blocked sensitive text.
func SensitiveContentPayload(err error) gin.H {
	h := gin.H{"error": "内容包含敏感词"}
	var sce *appmoderation.SensitiveContentError
	if errors.As(err, &sce) && len(sce.Hits) > 0 {
		h["sensitive_words"] = sce.Hits
	}
	return h
}
