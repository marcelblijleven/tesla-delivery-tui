package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/config"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
)

const (
	clientID            = "ownerapi"
	redirectURI         = "https://auth.tesla.com/void/callback"
	authURL             = "https://auth.tesla.com/oauth2/v3/authorize"
	tokenURL            = "https://auth.tesla.com/oauth2/v3/token"
	scope               = "openid email offline_access"
	codeChallengeMethod = "S256"
)

// AuthResult contains the result of an authentication attempt
type AuthResult struct {
	Tokens *model.TeslaTokens
	Error  error
}

// AuthSession holds the state for an auth flow
type AuthSession struct {
	CodeVerifier  string
	CodeChallenge string
	State         string
	AuthURL       string
}

// Auth handles Tesla OAuth2 authentication
type Auth struct {
	httpClient *http.Client
}

// NewAuth creates a new Auth instance
func NewAuth() *Auth {
	return &Auth{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// CreateAuthSession creates a new auth session with PKCE values
func (a *Auth) CreateAuthSession() (*AuthSession, error) {
	codeVerifier, err := config.GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeChallenge := config.GenerateCodeChallenge(codeVerifier)

	state, err := config.GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	params := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"response_type":         {"code"},
		"scope":                 {scope},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {codeChallengeMethod},
	}

	return &AuthSession{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		State:         state,
		AuthURL:       authURL + "?" + params.Encode(),
	}, nil
}

// ExchangeCode exchanges an authorization code for tokens (public method)
func (a *Auth) ExchangeCode(code, codeVerifier string) (*model.TeslaTokens, error) {
	return a.exchangeCodeForTokens(code, codeVerifier)
}

// exchangeCodeForTokens exchanges an authorization code for tokens
func (a *Auth) exchangeCodeForTokens(code, codeVerifier string) (*model.TeslaTokens, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token exchange failed: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	var tokens model.TeslaTokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Calculate expiry time
	tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	return &tokens, nil
}

// RefreshTokens uses the refresh token to get new tokens
func (a *Auth) RefreshTokens(refreshToken string) (*model.TeslaTokens, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token refresh failed: %s - %s", errResp.Error, errResp.ErrorDescription)
	}

	var tokens model.TeslaTokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	// Preserve the original refresh token if the response didn't include a new one
	if tokens.RefreshToken == "" {
		tokens.RefreshToken = refreshToken
	}

	// Calculate expiry time
	tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	return &tokens, nil
}
