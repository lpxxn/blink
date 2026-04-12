package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	appmoderation "github.com/lpxxn/blink/application/moderation"
	apppost "github.com/lpxxn/blink/application/post"
	apppostreply "github.com/lpxxn/blink/application/postreply"
	domainpost "github.com/lpxxn/blink/domain/post"
	domainpostreply "github.com/lpxxn/blink/domain/postreply"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	"github.com/gin-gonic/gin"
)

func (s *Server) ListReplies(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad post id"})
		return
	}
	var afterID *int64
	if v := c.Query("cursor"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad cursor"})
			return
		}
		afterID = &id
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := s.Replies.List(c.Request.Context(), postID, afterID, limit)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) || errors.Is(err, apppost.ErrNotVisible) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := s.repliesToJSON(c.Request.Context(), list)
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, RepliesPageJSON{Replies: out, NextCursor: next})
}

type createReplyBody struct {
	Body          string `json:"body"`
	ParentReplyID *int64 `json:"parent_reply_id,string"`
}

func (s *Server) CreateReply(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad post id"})
		return
	}
	var body createReplyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rep, err := s.Replies.Add(c.Request.Context(), uid, postID, body.Body, body.ParentReplyID)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) || errors.Is(err, apppost.ErrNotVisible) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, apppostreply.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, appmoderation.ErrSensitiveContent) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "内容包含敏感词"})
			return
		}
		if errors.Is(err, domainpostreply.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_reply_id"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s.replyToJSON(c.Request.Context(), rep))
}

func (s *Server) repliesToJSON(ctx context.Context, list []*domainpostreply.Reply) []ReplyJSON {
	if len(list) == 0 {
		return []ReplyJSON{}
	}
	ids := make([]int64, len(list))
	for i, r := range list {
		ids[i] = r.UserID
	}
	names := ResolveUserNames(ctx, s.Users, ids)
	out := make([]ReplyJSON, 0, len(list))
	for _, r := range list {
		j := ReplyToJSON(r)
		if names != nil {
			j.UserName = names[r.UserID]
		}
		out = append(out, j)
	}
	return out
}

func (s *Server) replyToJSON(ctx context.Context, r *domainpostreply.Reply) ReplyJSON {
	j := ReplyToJSON(r)
	if s.Users == nil {
		return j
	}
	names := ResolveUserNames(ctx, s.Users, []int64{r.UserID})
	if names != nil {
		j.UserName = names[r.UserID]
	}
	return j
}

func (s *Server) DeleteReply(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	err = s.Replies.DeleteOwn(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, domainpostreply.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, apppostreply.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
