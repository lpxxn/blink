package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appcategory "github.com/lpxxn/blink/application/category"
	appmoderation "github.com/lpxxn/blink/application/moderation"
	apppost "github.com/lpxxn/blink/application/post"
	apppostlike "github.com/lpxxn/blink/application/postlike"
	domainpost "github.com/lpxxn/blink/domain/post"
	domainuser "github.com/lpxxn/blink/domain/user"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) ListPosts(c *gin.Context) {
	var categoryID *int64
	if v := c.Query("category_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad category_id"})
			return
		}
		categoryID = &id
	}
	uncategorized := c.Query("uncategorized") == "1" || c.Query("uncategorized") == "true"
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
	list, err := s.Posts.ListFeed(c.Request.Context(), categoryID, uncategorized, beforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var viewer *int64
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
	}
	out := s.postsToJSON(c.Request.Context(), list, viewer)
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, PostsPageJSON{Posts: out, NextCursor: next})
}

func (s *Server) GetPost(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var viewer *int64
	viewerIsSuperAdmin := false
	if uid, ok := httpauth.UserIDFromContext(c); ok {
		viewer = &uid
		if s.Users != nil {
			if u, err := s.Users.GetByID(c.Request.Context(), uid); err == nil && u.Role == domainuser.RoleSuperAdmin {
				viewerIsSuperAdmin = true
			}
		}
	}
	p, err := s.Posts.GetForViewer(c.Request.Context(), id, viewer, viewerIsSuperAdmin)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, apppost.ErrNotVisible) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s.postToJSON(c.Request.Context(), p, viewer))
}

type createPostBody struct {
	Body       string   `json:"body"`
	CategoryID int64    `json:"category_id,string"`
	Images     []string `json:"images"`
	Draft      bool     `json:"draft"`
}

func (s *Server) CreatePost(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var body createPostBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Images == nil {
		body.Images = []string{}
	}
	p, err := s.Posts.Create(c.Request.Context(), uid, body.Body, &body.CategoryID, body.Images, body.Draft)
	if err != nil {
		if errors.Is(err, apppost.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, appcategory.ErrInvalidCategory) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		if errors.Is(err, appmoderation.ErrSensitiveContent) {
			c.JSON(http.StatusBadRequest, SensitiveContentPayload(err))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s.postToJSON(c.Request.Context(), p, &uid))
}

type patchPostBody struct {
	Body          *string   `json:"body"`
	CategoryID    *int64    `json:"category_id,string"`
	ClearCategory *bool     `json:"clear_category"`
	Images        *[]string `json:"images"`
	Status        *int      `json:"status"`
}

func (s *Server) PatchPost(c *gin.Context) {
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
	var body patchPostBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	patch := apppost.Patch{
		Body:          body.Body,
		CategoryID:    body.CategoryID,
		ClearCategory: body.ClearCategory != nil && *body.ClearCategory,
		Images:        body.Images,
		Status:        body.Status,
	}
	p, err := s.Posts.Patch(c.Request.Context(), uid, id, patch)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, apppost.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if errors.Is(err, apppost.ErrCannotPublishRemoved) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "帖子已被管理员下架，无法直接设为已发布；请先修改内容后「申请复核」或提交申诉。"})
			return
		}
		if errors.Is(err, apppost.ErrInvalidInput) || errors.Is(err, appcategory.ErrInvalidCategory) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, appmoderation.ErrSensitiveContent) {
			c.JSON(http.StatusBadRequest, SensitiveContentPayload(err))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s.postToJSON(c.Request.Context(), p, &uid))
}

func (s *Server) DeletePost(c *gin.Context) {
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
	err = s.Posts.Delete(c.Request.Context(), uid, id)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, apppost.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}

func (s *Server) ListMyPosts(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	includeDraft := c.Query("include_draft") == "1" || c.Query("include_draft") == "true"
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
	list, err := s.Posts.ListMine(c.Request.Context(), uid, includeDraft, beforeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := s.postsToJSON(c.Request.Context(), list, &uid)
	var next *string
	if len(list) > 0 {
		next = NextCursorString(list[len(list)-1].ID)
	}
	c.JSON(http.StatusOK, PostsPageJSON{Posts: out, NextCursor: next})
}

func (s *Server) postsToJSON(ctx context.Context, list []*domainpost.Post, viewerID *int64) []PostJSON {
	if len(list) == 0 {
		return []PostJSON{}
	}
	userIDs := make([]int64, len(list))
	postIDs := make([]int64, len(list))
	for i, p := range list {
		userIDs[i] = p.UserID
		postIDs[i] = p.ID
	}
	names := ResolveUserNames(ctx, s.Users, userIDs)
	likeMeta := s.likeMetaMap(ctx, postIDs, viewerID)
	out := make([]PostJSON, 0, len(list))
	for _, p := range list {
		j := PostToJSON(p)
		if names != nil {
			j.UserName = names[p.UserID]
		}
		if m, ok := likeMeta[p.ID]; ok {
			j.LikeCount = m.LikeCount
			if viewerID != nil {
				liked := m.Liked
				j.Liked = &liked
			}
		}
		out = append(out, j)
	}
	return out
}

func (s *Server) postToJSON(ctx context.Context, p *domainpost.Post, viewerID *int64) PostJSON {
	j := PostToJSON(p)
	if s.Users != nil {
		names := ResolveUserNames(ctx, s.Users, []int64{p.UserID})
		if names != nil {
			j.UserName = names[p.UserID]
		}
	}
	if m, ok := s.likeMetaMap(ctx, []int64{p.ID}, viewerID)[p.ID]; ok {
		j.LikeCount = m.LikeCount
		if viewerID != nil {
			liked := m.Liked
			j.Liked = &liked
		}
	}
	return j
}

func (s *Server) likeMetaMap(ctx context.Context, postIDs []int64, viewerID *int64) map[int64]apppostlike.PostLikeMeta {
	if s.Likes == nil || len(postIDs) == 0 {
		return nil
	}
	m, err := s.Likes.MetaForPosts(ctx, postIDs, viewerID)
	if err != nil {
		return nil
	}
	return m
}
