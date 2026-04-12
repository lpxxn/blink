package httpapi

import (
	appeventing "github.com/lpxxn/blink/application/eventing"
	appnotification "github.com/lpxxn/blink/application/notification"
	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	domaincategory "github.com/lpxxn/blink/domain/category"
	domainsession "github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// Server exposes JSON handlers under /api.
type Server struct {
	Posts         *apppost.Service
	Replies       *apppostreply.Service
	Notifications *appnotification.Service
	NotifyEvents  appeventing.NotificationPublisher
	Categories    domaincategory.Repository
	Users         domainuser.Repository
	Sessions      domainsession.Store
	UploadRoot    string
	UploadURLPath string
}
