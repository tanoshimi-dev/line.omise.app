package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/database"
)

// Health returns a Gin handler for GET /health. It reports process liveness
// unconditionally, plus best-effort database reachability via a real Ping
// against the pgx pool (see internal/database).
func Health(db *database.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbStatus := database.Check(c.Request.Context(), db)

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"database": gin.H{
				"configured": dbStatus.Configured,
				"reachable":  dbStatus.Reachable,
				"detail":     dbStatus.Detail,
			},
		})
	}
}
