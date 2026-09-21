package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/auth"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/middleware"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/repository"
	"github.com/tanoshimi-dev/line.omise.app/sys/02_backend/internal/session"
)

// oauthStateCookie holds "<state>.<nonce>" between the /login redirect and
// the provider's /callback, so the callback can verify both the CSRF state
// and (for providers that support it) the ID token's nonce.
const oauthStateCookie = "line_omise_oauth_state"
const oauthStateTTL = 10 * time.Minute

// AuthHandler implements dev-plan-04-auth's /auth/* endpoints.
type AuthHandler struct {
	Providers     map[string]auth.Provider
	Users         *repository.UserRepository
	Sessions      *repository.SessionRepository
	AdminEmails   []string
	SessionSecret string
	FrontendURL   string
	CookieSecure  bool
}

// Login redirects the browser to the named provider's authorization URL.
func (h *AuthHandler) Login(providerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider, ok := h.Providers[providerName]
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		state, err := session.RandomToken()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		nonce, err := session.RandomToken()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(oauthStateCookie, state+"."+nonce, int(oauthStateTTL.Seconds()), "/auth", "", h.CookieSecure, true)
		c.Redirect(http.StatusFound, provider.AuthURL(state, nonce))
	}
}

// Callback completes the authorization-code flow: verifies state/nonce,
// exchanges the code, upserts the user, and issues a session cookie.
func (h *AuthHandler) Callback(providerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider, ok := h.Providers[providerName]
		if !ok {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		code := c.Query("code")
		state := c.Query("state")
		if code == "" || state == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		raw, err := c.Cookie(oauthStateCookie)
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(oauthStateCookie, "", -1, "/auth", "", h.CookieSecure, true)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		expectedState, nonce, ok := strings.Cut(raw, ".")
		if !ok || expectedState != state {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		claims, err := provider.Exchange(c.Request.Context(), code, nonce)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		isAdmin := emailIn(claims.Email, h.AdminEmails)
		user, err := h.Users.UpsertByProvider(c.Request.Context(), claims.Provider, claims.ProviderUserID, claims.Email, claims.DisplayName, claims.AvatarURL, isAdmin)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		sess, err := h.Sessions.Create(c.Request.Context(), user.ID)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(session.CookieName, session.Sign(sess.ID, h.SessionSecret), int(repository.SessionTTL.Seconds()), "/", "", h.CookieSecure, true)
		c.Redirect(http.StatusFound, h.FrontendURL+"/auth/callback")
	}
}

// Logout destroys the current session (if any) and clears the cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	if raw, err := c.Cookie(session.CookieName); err == nil {
		if id, ok := session.Verify(raw, h.SessionSecret); ok {
			_ = h.Sessions.Delete(c.Request.Context(), id)
		}
	}
	h.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

// DeleteMe permanently deletes the authenticated user's account and all
// current user-owned data through the database's foreign-key cascades.
func (h *AuthHandler) DeleteMe(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if err := h.Users.DeleteByID(c.Request.Context(), user.ID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	h.clearSessionCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(session.CookieName, "", -1, "/", "", h.CookieSecure, true)
}

// Me returns the logged-in user. Mount behind middleware.RequireReader.
func (h *AuthHandler) Me(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":           strconv.FormatInt(user.ID, 10),
		"provider":     user.Provider,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"avatar_url":   user.AvatarURL,
		"role":         user.Role,
	})
}

func emailIn(email string, allowlist []string) bool {
	if email == "" {
		return false
	}
	for _, e := range allowlist {
		if strings.EqualFold(e, email) {
			return true
		}
	}
	return false
}
