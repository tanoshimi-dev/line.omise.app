package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	googleAuthURL      = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL     = "https://oauth2.googleapis.com/token"
	googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"
)

// GoogleClient implements Provider for Google OAuth 2.0 / OpenID Connect.
//
// Like LineClient, ID token verification uses Google's hosted tokeninfo
// endpoint (GET ?id_token=...) instead of a local JWKS-based JWT check.
// Google documents tokeninfo as valid for server-side verification; the
// audience and nonce are checked here against what we issued.
type GoogleClient struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
	HTTPClient   *http.Client
}

func NewGoogleClient(clientID, clientSecret, callbackURL string) *GoogleClient {
	return &GoogleClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackURL:  callbackURL,
		HTTPClient:   http.DefaultClient,
	}
}

func (c *GoogleClient) Name() string { return "google" }

func (c *GoogleClient) AuthURL(state, nonce string) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {c.ClientID},
		"redirect_uri":  {c.CallbackURL},
		"state":         {state},
		"nonce":         {nonce},
		"scope":         {"openid email profile"},
	}
	return googleAuthURL + "?" + q.Encode()
}

type googleTokenResponse struct {
	IDToken string `json:"id_token"`
}

type googleTokenInfo struct {
	Sub     string `json:"sub"`
	Aud     string `json:"aud"`
	Nonce   string `json:"nonce"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Error   string `json:"error_description"`
}

func (c *GoogleClient) Exchange(ctx context.Context, code, nonce string) (*ProviderClaims, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {c.CallbackURL},
		"client_id":     {c.ClientID},
		"client_secret": {c.ClientSecret},
	}
	tokenBody, err := c.postForm(ctx, googleTokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("google: token exchange: %w", err)
	}

	var token googleTokenResponse
	if err := json.Unmarshal(tokenBody, &token); err != nil {
		return nil, fmt.Errorf("google: decode token response: %w", err)
	}
	if token.IDToken == "" {
		return nil, fmt.Errorf("google: token response missing id_token")
	}

	infoBody, err := c.get(ctx, googleTokenInfoURL+"?id_token="+url.QueryEscape(token.IDToken))
	if err != nil {
		return nil, fmt.Errorf("google: id token verify: %w", err)
	}

	var claims googleTokenInfo
	if err := json.Unmarshal(infoBody, &claims); err != nil {
		return nil, fmt.Errorf("google: decode tokeninfo response: %w", err)
	}
	if claims.Error != "" {
		return nil, fmt.Errorf("google: id token invalid: %s", claims.Error)
	}
	if claims.Aud != c.ClientID {
		return nil, fmt.Errorf("google: id token audience mismatch")
	}
	if claims.Nonce != nonce {
		return nil, fmt.Errorf("google: id token nonce mismatch")
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("google: tokeninfo response missing sub")
	}

	return &ProviderClaims{
		Provider:       "google",
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		DisplayName:    claims.Name,
		AvatarURL:      claims.Picture,
	}, nil
}

func (c *GoogleClient) postForm(ctx context.Context, endpoint string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req)
}

func (c *GoogleClient) get(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *GoogleClient) do(req *http.Request) ([]byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}
