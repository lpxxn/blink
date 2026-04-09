package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListCategories(c *gin.Context) {
	if s.Categories == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "categories unavailable"})
		return
	}
	list, err := s.Categories.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]CategoryJSON, 0, len(list))
	for _, cat := range list {
		out = append(out, CategoryToJSON(cat))
	}
	c.JSON(http.StatusOK, CategoriesResponse{Categories: out})
}
