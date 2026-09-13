// Package config loads server configuration from environment variables.
package config

import (
	"os"
	"strings"
)

// Config holds the process-wide configuration loaded from the environment.
type Config struct {
	Port                string
	DatabaseURL         string
	CORSAllowedOrigins  []string
}

// Load reads configuration from environment variables, applying defaults
// for anything left unset.
func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		CORSAllowedOrigins: splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS")),
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
