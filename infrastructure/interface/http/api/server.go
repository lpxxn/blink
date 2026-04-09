package httpapi

import (
	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	domaincategory "github.com/lpxxn/blink/domain/category"
)

// Server exposes JSON handlers under /api.
type Server struct {
	Posts         *apppost.Service
	Replies       *apppostreply.Service
	Categories    domaincategory.Repository
	UploadRoot    string
	UploadURLPath string
}
