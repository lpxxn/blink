package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apppost "github.com/lpxxn/blink/application/post"
	domainpost "github.com/lpxxn/blink/domain/post"
	httpauth "github.com/lpxxn/blink/infrastructure/interface/http/auth"
)

type moderationRequestBody struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func (s *Server) SubmitModerationRequest(c *gin.Context) {
	uid, ok := httpauth.UserIDFromContext(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body moderationRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := s.Posts.SubmitModerationRequest(c.Request.Context(), uid, postID, body.Kind, body.Message)
	if err != nil {
		if errors.Is(err, domainpost.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if errors.Is(err, apppost.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if errors.Is(err, apppost.ErrAppealNotAllowed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅在下架状态下可提交申诉或复核申请"})
			return
		}
		if errors.Is(err, apppost.ErrAppealPending) {
			c.JSON(http.StatusConflict, gin.H{"error": "已有待处理的申诉/复核申请"})
			return
		}
		if errors.Is(err, apppost.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求无效：kind 须为 appeal 或 resubmit；申诉须填写说明"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s.postToJSON(c.Request.Context(), p))
}
