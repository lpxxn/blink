package httpadmin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (s *Server) ListAuditLogs(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, total, err := s.Admin.ListAuditLogs(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, e := range list {
		row := gin.H{
			"id":          strconv.FormatInt(e.ID, 10),
			"actor_id":    strconv.FormatInt(e.ActorID, 10),
			"action":      e.Action,
			"target_type": e.TargetType,
			"detail":      e.Detail,
			"created_at":  e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
		if e.TargetID != nil {
			row["target_id"] = strconv.FormatInt(*e.TargetID, 10)
		}
		out = append(out, row)
	}
	c.JSON(http.StatusOK, gin.H{"logs": out, "total": strconv.FormatInt(total, 10)})
}
