package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appfeedback "github.com/lpxxn/blink/application/feedback"
	domainfeedback "github.com/lpxxn/blink/domain/feedback"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

type feedbackBody struct {
	Body string `json:"body"`
}

func (s *Server) CreateFeedback(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Feedback == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	var body feedbackBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := s.Feedback.Create(c.Request.Context(), uid, body.Body)
	if err != nil {
		if errors.Is(err, appfeedback.ErrInvalidBody) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "反馈内容不能为空，且不能超过 4000 字符"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, FeedbackThreadToJSON(t))
}

func (s *Server) ListMyFeedback(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if s.Feedback == nil {
		c.JSON(http.StatusOK, FeedbackThreadPageJSON{Feedback: []FeedbackThreadJSON{}})
		return
	}
	var beforeID *int64
	if v := c.Query("cursor"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		beforeID = &id
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, err := s.Feedback.ListForUser(c.Request.Context(), uid, beforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]FeedbackThreadJSON, 0, len(list))
	for _, t := range list {
		out = append(out, FeedbackThreadToJSON(t))
	}
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, FeedbackThreadPageJSON{Feedback: out, NextCursor: next})
}

func (s *Server) GetMyFeedback(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
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
	t, msgs, err := s.Feedback.GetForUser(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, domainfeedback.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appfeedback.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, feedbackDetailToJSON(t, msgs))
}

func (s *Server) ReplyMyFeedback(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
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
	var body feedbackBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg, err := s.Feedback.ReplyByUser(c.Request.Context(), uid, id, body.Body)
	if err != nil {
		switch {
		case errors.Is(err, domainfeedback.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		case errors.Is(err, appfeedback.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, appfeedback.ErrReplyLimit):
			c.JSON(http.StatusConflict, gin.H{"error": "用户最多只能追加反馈 2 次"})
		case errors.Is(err, appfeedback.ErrInvalidBody):
			c.JSON(http.StatusBadRequest, gin.H{"error": "反馈内容不能为空，且不能超过 4000 字符"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, FeedbackMessageToJSON(msg))
}

func feedbackDetailToJSON(t *domainfeedback.Thread, msgs []*domainfeedback.Message) FeedbackDetailJSON {
	out := make([]FeedbackMessageJSON, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, FeedbackMessageToJSON(m))
	}
	return FeedbackDetailJSON{Thread: FeedbackThreadToJSON(t), Messages: out}
}
