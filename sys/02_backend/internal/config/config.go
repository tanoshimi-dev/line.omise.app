// Package config loads server configuration from environment variables.
package config

import (
	"os"
	"strings"
)

// Config holds the process-wide configuration loaded from the environment.
type Config struct {
	Port               string
	DatabaseURL        string
	CORSAllowedOrigins []string

	// Env is "development" (default) or "production". It only affects the
	// session cookie's Secure flag today (dev-plan-04-auth).
	Env string

	// FrontendURL is where the browser is redirected after a login/OAuth
	// callback completes (see internal/handler.AuthHandler). Not part of the
	// dev-plan-04-auth env var list, but required to close the login loop —
	// see the Step 04 result doc for why it was added.
	FrontendURL string

	LineLoginChannelID     string
	LineLoginChannelSecret string
	LineLoginCallbackURL   string

	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	GoogleOAuthCallbackURL  string

	// AdminEmails lists addresses that are auto-promoted to role=admin on
	// login (dev-plan-04-auth 4.5).
	AdminEmails []string

	// SessionSecret signs the session cookie (HMAC) so a tampered cookie
	// value is rejected before it ever reaches the database.
	SessionSecret string
}

// Load reads configuration from environment variables, applying defaults
// for anything left unset.
func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		CORSAllowedOrigins: splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
		Env:                getEnv("APP_ENV", "development"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),

		LineLoginChannelID:     os.Getenv("LINE_LOGIN_CHANNEL_ID"),
		LineLoginChannelSecret: os.Getenv("LINE_LOGIN_CHANNEL_SECRET"),
		LineLoginCallbackURL:   os.Getenv("LINE_LOGIN_CALLBACK_URL"),

		GoogleOAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GoogleOAuthCallbackURL:  os.Getenv("GOOGLE_OAUTH_CALLBACK_URL"),

		AdminEmails:   splitCSV(os.Getenv("ADMIN_EMAILS")),
		SessionSecret: os.Getenv("SESSION_SECRET"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
