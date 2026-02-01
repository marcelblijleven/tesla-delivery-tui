package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/config"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
)

// Client is the Tesla API client
type Client struct {
	httpClient *http.Client
	config     *config.Config
	auth       *Auth
	tokens     *model.TeslaTokens
	mu sync.Mutex // protects token refresh
}

// NewClient creates a new Tesla API client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		config:     cfg,
		auth:       NewAuth(),
	}
}

// SetTokens sets the current tokens
func (c *Client) SetTokens(tokens *model.TeslaTokens) {
	c.tokens = tokens
}

// GetTokens returns the current tokens
func (c *Client) GetTokens() *model.TeslaTokens {
	return c.tokens
}

// Auth returns the auth handler
func (c *Client) Auth() *Auth {
	return c.auth
}

// EnsureValidTokens ensures tokens are valid, refreshing if needed
func (c *Client) EnsureValidTokens() error {
	if c.tokens == nil {
		return fmt.Errorf("not authenticated")
	}

	if !c.tokens.IsExpired() {
		return nil
	}

	// Attempt to refresh tokens
	newTokens, err := c.auth.RefreshTokens(c.tokens.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh tokens: %w", err)
	}

	c.tokens = newTokens

	// Save the new tokens
	if err := c.config.SaveTokens(newTokens); err != nil {
		return fmt.Errorf("failed to save refreshed tokens: %w", err)
	}

	return nil
}

// doRequest performs an authenticated API request
func (c *Client) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	if err := c.EnsureValidTokens(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.tokens.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for auth errors
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		// Lock to prevent concurrent refresh attempts
		c.mu.Lock()
		newTokens, err := c.auth.RefreshTokens(c.tokens.RefreshToken)
		if err != nil {
			c.mu.Unlock()
			return nil, fmt.Errorf("token expired and refresh failed: %w", err)
		}

		c.tokens = newTokens
		if saveErr := c.config.SaveTokens(newTokens); saveErr != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to save refreshed tokens: %v\n", saveErr)
		}
		c.mu.Unlock()

		// Retry the request with new token
		req, err = http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create retry request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.tokens.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("retry request failed: %w", err)
		}
	}

	return resp, nil
}

// Get performs an authenticated GET request
func (c *Client) Get(url string) (*http.Response, error) {
	return c.doRequest("GET", url, nil)
}

// decodeResponse decodes a JSON response into the target
func decodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
