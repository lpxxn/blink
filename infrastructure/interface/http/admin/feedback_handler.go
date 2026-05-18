package httpadmin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appfeedback "github.com/lpxxn/blink/application/feedback"
	domainfeedback "github.com/lpxxn/blink/domain/feedback"
	httpapi "github.com/lpxxn/blink/infrastructure/interface/http/api"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) ListFeedback(c *gin.Context) {
	if s.Feedback == nil {
		c.JSON(http.StatusOK, httpapi.FeedbackThreadPageJSON{Feedback: []httpapi.FeedbackThreadJSON{}})
		return
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	var f domainfeedback.ListFilters
	if v := c.Query("status"); v != "" {
		f.Status = &v
	}
	if v := c.Query("user_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad user_id"})
			return
		}
		f.UserID = &id
	}
	list, total, err := s.Feedback.ListForAdmin(c.Request.Context(), f, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]httpapi.FeedbackThreadJSON, 0, len(list))
	ids := make([]int64, 0, len(list))
	for _, t := range list {
		out = append(out, httpapi.FeedbackThreadToJSON(t))
		ids = append(ids, t.UserID)
	}
	if names := httpapi.ResolveUserNames(c.Request.Context(), s.Users, ids); names != nil {
		for i := range out {
			out[i].UserName = names[out[i].UserID]
		}
	}
	c.JSON(http.StatusOK, httpapi.FeedbackThreadPageJSON{Feedback: out, Total: total})
}

func (s *Server) GetFeedback(c *gin.Context) {
	if s.Feedback == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	t, msgs, err := s.Feedback.GetForAdmin(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domainfeedback.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	d := httpapi.FeedbackDetailJSON{
		Thread:   httpapi.FeedbackThreadToJSON(t),
		Messages: make([]httpapi.FeedbackMessageJSON, 0, len(msgs)),
	}
	if names := httpapi.ResolveUserNames(c.Request.Context(), s.Users, []int64{t.UserID}); names != nil {
		d.Thread.UserName = names[t.UserID]
	}
	for _, m := range msgs {
		d.Messages = append(d.Messages, httpapi.FeedbackMessageToJSON(m))
	}
	c.JSON(http.StatusOK, d)
}

type adminFeedbackReplyBody struct {
	Body string `json:"body"`
}

func (s *Server) ReplyFeedback(c *gin.Context) {
	adminID, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Feedback == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body adminFeedbackReplyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg, err := s.Feedback.ReplyByAdmin(c.Request.Context(), adminID, id, body.Body)
	if err != nil {
		switch {
		case errors.Is(err, domainfeedback.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case errors.Is(err, appfeedback.ErrInvalidBody):
			c.JSON(http.StatusBadRequest, gin.H{"error": "回复内容不能为空，且不能超过 4000 字符"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, httpapi.FeedbackMessageToJSON(msg))
}

func (s *Server) CloseFeedback(c *gin.Context) {
	actorID, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := s.Admin.CloseFeedback(c.Request.Context(), actorID, id); err != nil {
		if errors.Is(err, domainfeedback.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
