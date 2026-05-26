package httpapi

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	"github.com/lpxxn/blink/infrastructure/sse"
)

// NotificationStream is an SSE endpoint: GET /api/me/notifications/stream.
// It keeps the connection open and pushes events whenever a new notification
// is created for the authenticated user.
func (s *Server) NotificationStream(c *gin.Context) {
	if s.SSEHub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SSE not available"})
		return
	}
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	ch := s.SSEHub.Subscribe(uid)
	defer s.SSEHub.Unsubscribe(uid, ch)

	// Send an initial "connected" event so the client knows the stream is live.
	fmt.Fprintf(c.Writer, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	ctx := c.Request.Context()
	c.Stream(func(_ io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case evt, ok := <-ch:
			if !ok {
				return false
			}
			c.SSEvent(evt.Name, evt.Data)
			flusher.Flush()
			return true
		}
	})
}

// PublishNotificationEvent pushes a lightweight SSE event to the user so the
// browser can update the unread badge without polling.
func PublishNotificationEvent(hub *sse.Hub, userID int64, unreadCount int64) {
	if hub == nil || userID == 0 {
		return
	}
	data := fmt.Sprintf(`{"unread_count":"%d"}`, unreadCount)
	hub.Publish(userID, sse.Event{Name: "notification", Data: data})
}
