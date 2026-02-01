package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/config"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
	"github.com/pkg/browser"
)

const (
	clientID            = "ownerapi"
	redirectURI         = "https://auth.tesla.com/void/callback"
	authURL             = "https://auth.tesla.com/oauth2/v3/authorize"
	tokenURL            = "https://auth.tesla.com/oauth2/v3/token"
	scope               = "openid email offline_access"
	codeChallengeMethod = "S256"
	callbackServerAddr  = "127.0.0.1:8854"
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

// StartAuthFlow initiates the OAuth2 PKCE flow
// Returns a channel that will receive the auth result
func (a *Auth) StartAuthFlow(ctx context.Context) (<-chan AuthResult, error) {
	session, err := a.CreateAuthSession()
	if err != nil {
		return nil, err
	}

	resultChan := make(chan AuthResult, 1)

	// Start local callback server
	codeChan := make(chan string, 1)
	server := a.startCallbackServer(session.State, codeChan)

	// Open browser
	if err := browser.OpenURL(session.AuthURL); err != nil {
		server.Close()
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	// Wait for callback in goroutine
	go func() {
		defer server.Close()

		select {
		case code := <-codeChan:
			if code == "" {
				resultChan <- AuthResult{Error: fmt.Errorf("authentication cancelled or failed")}
				return
			}

			tokens, err := a.exchangeCodeForTokens(code, session.CodeVerifier)
			if err != nil {
				resultChan <- AuthResult{Error: err}
				return
			}
			resultChan <- AuthResult{Tokens: tokens}

		case <-ctx.Done():
			resultChan <- AuthResult{Error: ctx.Err()}

		case <-time.After(5 * time.Minute):
			resultChan <- AuthResult{Error: fmt.Errorf("authentication timeout")}
		}
	}()

	return resultChan, nil
}

// startCallbackServer starts a local HTTP server to receive the OAuth callback
func (a *Auth) startCallbackServer(expectedState string, codeChan chan<- string) *http.Server {
	mux := http.NewServeMux()

	// Main page - serves the callback handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Tesla TUI - Authentication</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            background: #1a1a1a;
            color: #fff;
        }
        h1 { color: #e31937; }
        .instructions {
            background: #2a2a2a;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        input[type="text"] {
            width: 100%;
            padding: 12px;
            margin: 10px 0;
            border: 1px solid #444;
            border-radius: 4px;
            background: #333;
            color: #fff;
            font-size: 14px;
        }
        button {
            background: #e31937;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover { background: #c41730; }
        .success { color: #22c55e; }
        .error { color: #ef4444; }
        code {
            background: #333;
            padding: 2px 6px;
            border-radius: 3px;
        }
    </style>
</head>
<body>
    <h1>⚡ Tesla TUI Authentication</h1>

    <div class="instructions">
        <h3>Instructions:</h3>
        <ol>
            <li>Complete the Tesla login in the browser window that opened</li>
            <li>After login, you'll be redirected to a page that says "Page Not Found"</li>
            <li>Copy the <strong>entire URL</strong> from your browser's address bar</li>
            <li>Paste it below and click Submit</li>
        </ol>
        <p>The URL should look like: <code>https://auth.tesla.com/void/callback?code=...</code></p>
    </div>

    <form id="authForm">
        <input type="text" id="callbackUrl" placeholder="Paste the callback URL here..." autofocus />
        <button type="submit">Submit</button>
    </form>

    <p id="status"></p>

    <script>
        // Check if we have a code in the current URL (in case of redirect)
        const currentUrl = window.location.href;
        const urlParams = new URLSearchParams(window.location.search);
        const hashParams = new URLSearchParams(window.location.hash.substring(1));

        let code = urlParams.get('code') || hashParams.get('code');
        let state = urlParams.get('state') || hashParams.get('state');

        if (code) {
            submitCode(code, state);
        }

        document.getElementById('authForm').addEventListener('submit', function(e) {
            e.preventDefault();
            const url = document.getElementById('callbackUrl').value;

            try {
                const parsed = new URL(url);
                const params = new URLSearchParams(parsed.search);
                const hash = new URLSearchParams(parsed.hash.substring(1));

                code = params.get('code') || hash.get('code');
                state = params.get('state') || hash.get('state');

                if (!code) {
                    document.getElementById('status').innerHTML = '<span class="error">Could not find authorization code in URL</span>';
                    return;
                }

                submitCode(code, state);
            } catch (err) {
                document.getElementById('status').innerHTML = '<span class="error">Invalid URL format</span>';
            }
        });

        function submitCode(code, state) {
            fetch('/callback?code=' + encodeURIComponent(code) + '&state=' + encodeURIComponent(state || ''))
                .then(response => response.text())
                .then(data => {
                    document.getElementById('status').innerHTML = '<span class="success">✓ Authentication successful! You can close this window.</span>';
                    document.getElementById('authForm').style.display = 'none';
                })
                .catch(err => {
                    document.getElementById('status').innerHTML = '<span class="error">Error: ' + err.message + '</span>';
                });
        }
    </script>
</body>
</html>`)
	})

	// Callback endpoint
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h1>Error</h1><p>No authorization code provided.</p></body></html>`)
			return
		}

		// Validate state if provided
		if state != "" && state != expectedState {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h1>Error</h1><p>Invalid state parameter.</p></body></html>`)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body><h1>Success!</h1><p>You can close this window and return to the terminal.</p></body></html>`)

		// Send code non-blocking
		select {
		case codeChan <- code:
		default:
		}
	})

	server := &http.Server{
		Addr:    callbackServerAddr,
		Handler: mux,
	}

	go server.ListenAndServe()

	return server
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
