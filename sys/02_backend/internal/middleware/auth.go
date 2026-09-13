package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/session"
)

const contextUserKey = "auth.user"

// resolveUser reads and verifies the session cookie, returning the logged-in
// user or nil. It never calls c.Next()/c.Abort* itself — callers decide how
// to react — so it's safe to call from multiple middleware without each one
// accidentally cascading into the rest of the handler chain (RequireReader's
// own c.Next(), called this way, would otherwise invoke the downstream route
// handler before RequireAdmin got to check the role).
func resolveUser(c *gin.Context, sessions *repository.SessionRepository, users *repository.UserRepository, sessionSecret string) *repository.User {
	raw, err := c.Cookie(session.CookieName)
	if err != nil {
		return nil
	}

	id, ok := session.Verify(raw, sessionSecret)
	if !ok {
		return nil
	}

	sess, err := sessions.GetValid(c.Request.Context(), id)
	if err != nil {
		return nil
	}

	user, err := users.GetByID(c.Request.Context(), sess.UserID)
	if err != nil {
		return nil
	}
	return user
}

// RequireReader rejects the request with 401 unless it carries a valid
// session cookie, and attaches the resolved user to the Gin context for
// handlers to read via CurrentUser.
func RequireReader(sessions *repository.SessionRepository, users *repository.UserRepository, sessionSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := resolveUser(c, sessions, users, sessionSecret)
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(contextUserKey, user)
		c.Next()
	}
}

// RequireAdmin rejects the request with 401 (not logged in) or 403 (logged
// in but role != admin), and otherwise behaves like RequireReader.
func RequireAdmin(sessions *repository.SessionRepository, users *repository.UserRepository, sessionSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := resolveUser(c, sessions, users, sessionSecret)
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !user.IsAdmin() {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Set(contextUserKey, user)
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
