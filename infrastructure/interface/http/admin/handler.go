package httpadmin

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	appadmin "github.com/lpxxn/blink/application/admin"
	domainpost "github.com/lpxxn/blink/domain/post"
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
