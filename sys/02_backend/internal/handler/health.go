package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/database"
)

// Health returns a Gin handler for GET /health. It reports process liveness
// unconditionally, plus best-effort database reachability (see
// internal/database — the real driver arrives in dev-plan-02-database).
func Health(databaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbStatus := database.Check(c.Request.Context(), databaseURL)

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
