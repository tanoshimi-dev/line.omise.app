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
	lineAuthURL   = "https://access.line.me/oauth2/v2.1/authorize"
	lineTokenURL  = "https://api.line.me/oauth2/v2.1/token"
	lineVerifyURL = "https://api.line.me/oauth2/v2.1/verify"
)

// LineClient implements Provider for LINE Login.
//
// ID token verification uses LINE's hosted /oauth2/v2.1/verify endpoint
// (POST id_token + client_id [+ nonce]) rather than fetching LINE's JWKS and
// checking the JWT signature locally — LINE documents this endpoint as the
// supported server-side verification method, and it also checks nonce for
// us, so no extra JWT library is needed.
type LineClient struct {
	ChannelID     string
	ChannelSecret string
	CallbackURL   string
	HTTPClient    *http.Client
}

func NewLineClient(channelID, channelSecret, callbackURL string) *LineClient {
	return &LineClient{
		ChannelID:     channelID,
		ChannelSecret: channelSecret,
		CallbackURL:   callbackURL,
		HTTPClient:    http.DefaultClient,
	}
}

func (c *LineClient) Name() string { return "line" }

func (c *LineClient) AuthURL(state, nonce string) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {c.ChannelID},
		"redirect_uri":  {c.CallbackURL},
		"state":         {state},
		"nonce":         {nonce},
		"scope":         {"profile openid email"},
	}
	return lineAuthURL + "?" + q.Encode()
}

type lineTokenResponse struct {
	IDToken string `json:"id_token"`
}

type lineVerifyResponse struct {
	Sub              string `json:"sub"`
	Name             string `json:"name"`
	Picture          string `json:"picture"`
	Email            string `json:"email"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (c *LineClient) Exchange(ctx context.Context, code, nonce string) (*ProviderClaims, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {c.CallbackURL},
		"client_id":     {c.ChannelID},
		"client_secret": {c.ChannelSecret},
	}
	tokenResp, err := c.postForm(ctx, lineTokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("line: token exchange: %w", err)
	}

	var token lineTokenResponse
	if err := json.Unmarshal(tokenResp, &token); err != nil {
		return nil, fmt.Errorf("line: decode token response: %w", err)
	}
	if token.IDToken == "" {
		return nil, fmt.Errorf("line: token response missing id_token")
	}

	verifyForm := url.Values{
		"id_token":  {token.IDToken},
		"client_id": {c.ChannelID},
		"nonce":     {nonce},
	}
	verifyResp, err := c.postForm(ctx, lineVerifyURL, verifyForm)
	if err != nil {
		return nil, fmt.Errorf("line: id token verify: %w", err)
	}

	var claims lineVerifyResponse
	if err := json.Unmarshal(verifyResp, &claims); err != nil {
		return nil, fmt.Errorf("line: decode verify response: %w", err)
	}
	if claims.Error != "" {
		return nil, fmt.Errorf("line: id token invalid: %s (%s)", claims.Error, claims.ErrorDescription)
	}
	if claims.Sub == "" {
		return nil, fmt.Errorf("line: verify response missing sub")
	}

	return &ProviderClaims{
		Provider:       "line",
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		DisplayName:    claims.Name,
		AvatarURL:      claims.Picture,
	}, nil
}

func (c *LineClient) postForm(ctx context.Context, endpoint string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
