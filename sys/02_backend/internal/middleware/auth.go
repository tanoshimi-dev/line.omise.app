package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/session"
)

const contextUserKey = "auth.user"

// RequireReader rejects the request with 401 unless it carries a valid
// session cookie, and attaches the resolved user to the Gin context for
// handlers to read via CurrentUser.
func RequireReader(sessions *repository.SessionRepository, users *repository.UserRepository, sessionSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(session.CookieName)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		id, ok := session.Verify(raw, sessionSecret)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		sess, err := sessions.GetValid(c.Request.Context(), id)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		user, err := users.GetByID(c.Request.Context(), sess.UserID)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(contextUserKey, user)
		c.Next()
	}
}

// RequireAdmin builds on RequireReader, additionally rejecting non-admins
// with 403.
func RequireAdmin(sessions *repository.SessionRepository, users *repository.UserRepository, sessionSecret string) gin.HandlerFunc {
	requireReader := RequireReader(sessions, users, sessionSecret)
	return func(c *gin.Context) {
		requireReader(c)
		if c.IsAborted() {
			return
		}

		user := CurrentUser(c)
		if user == nil || !user.IsAdmin() {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

// CurrentUser returns the user attached by RequireReader/RequireAdmin, or
// nil outside a protected route.
func CurrentUser(c *gin.Context) *repository.User {
	v, ok := c.Get(contextUserKey)
	if !ok {
		return nil
	}
	user, _ := v.(*repository.User)
	return user
}
