package httpapi

import (
	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	domaincategory "github.com/lpxxn/blink/domain/category"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// Server exposes JSON handlers under /api.
type Server struct {
	Posts         *apppost.Service
	Replies       *apppostreply.Service
	Categories    domaincategory.Repository
	Users         domainuser.Repository
	UploadRoot    string
	UploadURLPath string
}
