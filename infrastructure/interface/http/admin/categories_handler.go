package httpadmin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appadmin "github.com/lpxxn/blink/application/admin"
	domaincategory "github.com/lpxxn/blink/domain/category"
	httpapi "github.com/lpxxn/blink/infrastructure/interface/http/api"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

func (s *Server) ListCategories(c *gin.Context) {
	list, err := s.Admin.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]httpapi.CategoryJSON, 0, len(list))
	for _, cat := range list {
		out = append(out, httpapi.CategoryToJSON(cat))
	}
	c.JSON(http.StatusOK, gin.H{"categories": out})
}

type createCategoryBody struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

func (s *Server) CreateCategory(c *gin.Context) {
	actorID, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var body createCategoryBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := s.Admin.CreateCategory(c.Request.Context(), actorID, body.Slug, body.Name, body.SortOrder)
	if err != nil {
		if errors.Is(err, appadmin.ErrInvalidCategory) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domaincategory.ErrDuplicate) {
			c.JSON(http.StatusConflict, gin.H{"error": "duplicate slug"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, httpapi.CategoryToJSON(cat))
}

type patchCategoryBody struct {
	Slug      *string `json:"slug"`
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
}

func (s *Server) PatchCategory(c *gin.Context) {
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
	var body patchCategoryBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := s.Admin.UpdateCategory(c.Request.Context(), actorID, id, body.Slug, body.Name, body.SortOrder)
	if err != nil {
		if errors.Is(err, domaincategory.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, appadmin.ErrInvalidCategory) || errors.Is(err, domaincategory.ErrDuplicate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, httpapi.CategoryToJSON(cat))
}

func (s *Server) DeleteCategory(c *gin.Context) {
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
	if err := s.Admin.DeleteCategory(c.Request.Context(), actorID, id); err != nil {
		if errors.Is(err, domaincategory.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
