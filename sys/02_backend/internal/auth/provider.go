// Package auth implements the LINE Login / Google OAuth authorization-code
// flows (dev-plan-04-auth). The frontend never sees a provider token — it
// only receives the authorization code redirect, and this package performs
// the token exchange and ID token verification server-side.
package auth

import "context"

// ProviderClaims is the normalized identity returned by a successful
// Exchange, regardless of provider.
type ProviderClaims struct {
	Provider       string // "line" | "google"
	ProviderUserID string
	Email          string
	DisplayName    string
	AvatarURL      string
}

// Provider is implemented by LineClient and GoogleClient.
type Provider interface {
	// Name returns the provider identifier used in URLs and ProviderClaims
	// ("line" or "google").
	Name() string

	// AuthURL builds the provider's authorization URL. state is an opaque
	// CSRF token the caller must verify on callback; nonce is included in
	// the request and re-checked against the returned ID token to prevent
	// replay.
	AuthURL(state, nonce string) string

	// Exchange trades an authorization code for the caller's identity,
	// verifying the ID token's signature, audience and nonce as part of
	// the provider's own verification endpoint.
	Exchange(ctx context.Context, code, nonce string) (*ProviderClaims, error)
}
