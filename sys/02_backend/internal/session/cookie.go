// Package session provides the signed session-cookie encoding shared by
// internal/handler (issues it on login) and internal/middleware (verifies it
// on every protected request). The session id itself lives in the `sessions`
// table (dev-plan-04-auth 4.3); the HMAC signature here only lets the server
// reject an obviously tampered cookie value without a database round trip.
package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

// CookieName is the browser cookie holding the signed session id.
const CookieName = "line_omise_session"

// RandomToken returns a cryptographically random, URL-safe token — used for
// session ids as well as the OAuth state/nonce values.
func RandomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// Sign returns "<id>.<hmac>" for storing in the session cookie.
func Sign(id, secret string) string {
	return id + "." + mac(id, secret)
}

// Verify checks a cookie value's signature and returns the embedded session
// id if valid.
func Verify(cookieValue, secret string) (id string, ok bool) {
	i := strings.LastIndex(cookieValue, ".")
	if i < 0 {
		return "", false
	}
	id, sig := cookieValue[:i], cookieValue[i+1:]
	expected := mac(id, secret)
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", false
	}
	return id, true
}

func mac(id, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(id))
	return hex.EncodeToString(h.Sum(nil))
}
