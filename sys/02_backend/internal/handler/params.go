package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseIDParam reads an int64 path param, aborting the request with 400 and
// returning ok=false if it's missing/invalid.
func parseIDParam(c *gin.Context, name string) (id int64, ok bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}
