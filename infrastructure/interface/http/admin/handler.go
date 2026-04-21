package httpadmin

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	appadmin "github.com/lpxxn/blink/application/admin"
	domainpost "github.com/lpxxn/blink/domain/post"
	domainpostreply "github.com/lpxxn/blink/domain/postreply"
	domainsensitiveword "github.com/lpxxn/blink/domain/sensitiveword"
	domainuser "github.com/lpxxn/blink/domain/user"
	httpapi "github.com/lpxxn/blink/infrastructure/interface/http/api"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
	"github.com/gin-gonic/gin"
)

// Server exposes /admin/api JSON handlers.
type Server struct {
	Admin         *appadmin.Service
	CategoryCount func(context.Context) (int64, error)
	Users         domainuser.Repository
	SMTP          *appadmin.SMTPSettings
}

func (s *Server) Overview(c *gin.Context) {
	o, err := s.Admin.Overview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ov := httpapi.OverviewJSON{
		UserCount:   o.UserCount,
		PostCount:   o.PostCount,
		PostsToday:  o.PostsToday,
	}
	if s.CategoryCount != nil {
		if n, err := s.CategoryCount(c.Request.Context()); err == nil {
			nv := n
			ov.CategoryCount = &nv
		}
	}
	c.JSON(http.StatusOK, ov)
}

func (s *Server) ListUsers(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := s.Admin.ListUsers(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]httpapi.AdminUserJSON, 0, len(list))
	for _, e := range list {
		out = append(out, httpapi.AdminUserToJSON(e))
	}
	c.JSON(http.StatusOK, httpapi.AdminUsersResponse{Users: out})
}

type patchUserBody struct {
	Status *int    `json:"status"`
	Role   *string `json:"role"`
}

func (s *Server) PatchUser(c *gin.Context) {
	actorID, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body patchUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.Admin.PatchUser(c.Request.Context(), actorID, targetID, body.Status, body.Role)
	if err != nil {
		if errors.Is(err, appadmin.ErrCannotDemoteSelf) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, appadmin.ErrInvalidRole) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

type resetPasswordBody struct {
	Password string `json:"password"`
}

func (s *Server) ResetUserPassword(c *gin.Context) {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body resetPasswordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.Admin.ResetUserPassword(c.Request.Context(), targetID, body.Password)
	if err != nil {
		if errors.Is(err, appadmin.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domainuser.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) ListPosts(c *gin.Context) {
	var f domainpost.AdminListFilters
	if v := c.Query("user_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad user_id"})
			return
		}
		f.UserID = &id
	}
	if v := c.Query("category_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad category_id"})
			return
		}
		f.CategoryID = &id
	}
	if v := c.Query("moderation_flag"); v != "" {
		mf, err := strconv.Atoi(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad moderation_flag"})
			return
		}
		f.ModerationFlag = &mf
	}
	if c.Query("include_deleted") == "1" || c.Query("include_deleted") == "true" {
		f.IncludeDeleted = true
	}
	if c.Query("appeal_pending") == "1" || c.Query("appeal_pending") == "true" {
		f.AppealPending = true
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, total, err := s.Admin.ListPosts(c.Request.Context(), f, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]httpapi.AdminPostJSON, 0, len(list))
	ids := make([]int64, 0, len(list))
	for _, p := range list {
		out = append(out, httpapi.AdminPostToJSON(p))
		ids = append(ids, p.UserID)
	}
	if names := httpapi.ResolveUserNames(c.Request.Context(), s.Users, ids); names != nil {
		for i := range out {
			out[i].UserName = names[out[i].UserID]
		}
	}
	c.JSON(http.StatusOK, httpapi.AdminPostsPageJSON{Posts: out, Total: total})
}

type patchAdminPostBody struct {
	ModerationFlag *int    `json:"moderation_flag"`
	ModerationNote *string `json:"moderation_note"`
	Status         *int    `json:"status"`
}

func (s *Server) PatchPost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body patchAdminPostBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := s.Admin.PatchPost(c.Request.Context(), id, body.ModerationFlag, body.ModerationNote, body.Status)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appadmin.ErrInvalidPostStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, appadmin.ErrInvalidModeration) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adm := httpapi.AdminPostToJSON(p)
	if names := httpapi.ResolveUserNames(c.Request.Context(), s.Users, []int64{p.UserID}); names != nil {
		adm.UserName = names[p.UserID]
	}
	c.JSON(http.StatusOK, adm)
}

type resolveAppealBody struct {
	Approve bool   `json:"approve"`
	Note    string `json:"note"`
}

func (s *Server) ResolveAppeal(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body resolveAppealBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := s.Admin.ResolveAppeal(c.Request.Context(), id, body.Approve, body.Note)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appadmin.ErrNoPendingAppeal) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	adm := httpapi.AdminPostToJSON(p)
	if names := httpapi.ResolveUserNames(c.Request.Context(), s.Users, []int64{p.UserID}); names != nil {
		adm.UserName = names[p.UserID]
	}
	c.JSON(http.StatusOK, adm)
}

func (s *Server) ListSensitiveWords(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, total, err := s.Admin.ListSensitiveWords(c.Request.Context(), offset, limit)
	if err != nil {
		if errors.Is(err, appadmin.ErrInvalidSensitiveWord) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sensitive words not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]httpapi.AdminSensitiveWordJSON, 0, len(list))
	for _, w := range list {
		out = append(out, httpapi.AdminSensitiveWordToJSON(w))
	}
	c.JSON(http.StatusOK, httpapi.AdminSensitiveWordsPageJSON{Words: out, Total: total})
}

type createSensitiveWordBody struct {
	Word string `json:"word"`
}

func (s *Server) CreateSensitiveWord(c *gin.Context) {
	var body createSensitiveWordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := s.Admin.CreateSensitiveWord(c.Request.Context(), body.Word)
	if err != nil {
		if errors.Is(err, appadmin.ErrInvalidSensitiveWord) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domainsensitiveword.ErrDuplicateWord) {
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate word"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, httpapi.AdminSensitiveWordToJSON(w))
}

type patchSensitiveWordBody struct {
	Enabled *bool `json:"enabled"`
}

func (s *Server) PatchSensitiveWord(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body patchSensitiveWordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w, err := s.Admin.PatchSensitiveWord(c.Request.Context(), id, body.Enabled)
	if err != nil {
		if errors.Is(err, domainsensitiveword.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appadmin.ErrInvalidSensitiveWord) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sensitive words not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpapi.AdminSensitiveWordToJSON(w))
}

func (s *Server) DeleteSensitiveWord(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	err = s.Admin.DeleteSensitiveWord(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domainsensitiveword.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appadmin.ErrInvalidSensitiveWord) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sensitive words not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

type patchAdminReplyBody struct {
	Hidden *bool `json:"hidden"`
}

func (s *Server) PatchReply(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body patchAdminReplyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Hidden == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing hidden"})
		return
	}
	if *body.Hidden {
		err = s.Admin.HidePostReply(c.Request.Context(), id)
	} else {
		err = s.Admin.UnhidePostReply(c.Request.Context(), id)
	}
	if err != nil {
		if errors.Is(err, domainpostreply.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appadmin.ErrRepliesNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "replies not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) ListPostReplies(c *gin.Context) {
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

	list, err := s.Admin.ListPostReplies(c.Request.Context(), postID, afterID, limit)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appadmin.ErrRepliesNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "replies not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]httpapi.ReplyJSON, 0, len(list))
	for _, r := range list {
		out = append(out, httpapi.ReplyToJSON(r))
	}
	// Resolve user names in batch (optional).
	if s.Users != nil && len(out) > 0 {
		ids := make([]int64, 0, len(out))
		for _, r := range out {
			ids = append(ids, r.UserID)
		}
		names := httpapi.ResolveUserNames(c.Request.Context(), s.Users, ids)
		if names != nil {
			for i := range out {
				out[i].UserName = names[out[i].UserID]
			}
		}
	}

	var next *string
	if len(list) > 0 {
		next = httpapi.NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, httpapi.RepliesPageJSON{Replies: out, NextCursor: next})
}

type sensitivePostModeJSON struct {
	Mode string `json:"mode"`
}

func (s *Server) GetSensitivePostMode(c *gin.Context) {
	mode, err := s.Admin.GetSensitivePostMode(c.Request.Context())
	if err != nil {
		if errors.Is(err, appadmin.ErrSettingsNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensitivePostModeJSON{Mode: mode})
}

func (s *Server) SetSensitivePostMode(c *gin.Context) {
	var body sensitivePostModeJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.Admin.SetSensitivePostMode(c.Request.Context(), body.Mode); err != nil {
		if errors.Is(err, appadmin.ErrSettingsNotConfigured) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "settings not configured"})
			return
		}
		if errors.Is(err, appadmin.ErrInvalidSetting) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mode"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
