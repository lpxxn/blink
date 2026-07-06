package httpapi

import (
	appauth "github.com/lpxxn/blink/application/auth"
	appeventing "github.com/lpxxn/blink/application/eventing"
	appfollow "github.com/lpxxn/blink/application/follow"
	appfeedback "github.com/lpxxn/blink/application/feedback"
	appblock "github.com/lpxxn/blink/application/block"
	apppostlike "github.com/lpxxn/blink/application/postlike"
	appreplylike "github.com/lpxxn/blink/application/replylike"
	appnotification "github.com/lpxxn/blink/application/notification"
	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	appsearch "github.com/lpxxn/blink/application/search"
	domaincategory "github.com/lpxxn/blink/domain/category"
	domainsession "github.com/lpxxn/blink/domain/session"
	domainuser "github.com/lpxxn/blink/domain/user"
	"github.com/lpxxn/blink/infrastructure/sse"
)

// Server exposes JSON handlers under /api.
type Server struct {
	Posts         *apppost.Service
	Replies       *apppostreply.Service
	Search        *appsearch.Service
	Follows       *appfollow.Service
	Likes         *apppostlike.Service
	ReplyLikes    *appreplylike.Service
	Blocks        *appblock.Service
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
