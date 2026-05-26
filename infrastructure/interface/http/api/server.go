package httpapi

import (
	appauth "github.com/lpxxn/blink/application/auth"
	appeventing "github.com/lpxxn/blink/application/eventing"
	appfollow "github.com/lpxxn/blink/application/follow"
	appfeedback "github.com/lpxxn/blink/application/feedback"
	apppostlike "github.com/lpxxn/blink/application/postlike"
	appnotification "github.com/lpxxn/blink/application/notification"
	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	domaincategory "github.com/lpxxn/blink/domain/category"
	domainsession "github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/infrastructure/sse"
)

// Server exposes JSON handlers under /api.
type Server struct {
	Posts         *apppost.Service
	Replies       *apppostreply.Service
	Follows       *appfollow.Service
	Likes         *apppostlike.Service
	Notifications *appnotification.Service
	Feedback      *appfeedback.Service
	NotifyEvents  appeventing.NotificationPublisher
	Categories    domaincategory.Repository
	Users         domainuser.Repository
	Sessions      domainsession.Store
	Passwords     *appauth.PasswordService
	UploadRoot    string
	UploadURLPath string
	SSEHub        *sse.Hub
}
