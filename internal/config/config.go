package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
	"github.com/zalando/go-keyring"
)

const (
	appName       = "tesla-delivery-tui"
	configDirName = ".config"
	tokensFile    = "tokens.enc"
	keyFile       = "key"

	// Keyring identifiers
	keyringService = "tesla-delivery-tui"
	keyringUser    = "tokens"
)

// Config holds application configuration
type Config struct {
	configDir       string
	keyringAvailable bool
}

// New creates a new Config instance
func New() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, configDirName, appName)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	c := &Config{configDir: configDir}

	// Test if keyring is available
	c.keyringAvailable = c.testKeyring()

	return c, nil
}

// testKeyring checks if the system keyring is available
func (c *Config) testKeyring() bool {
	// Try to access keyring with a test operation
	testKey := keyringService + "-test"

	// Try to set a test value
	err := keyring.Set(testKey, "test", "test")
	if err != nil {
		return false
	}

	// Clean up test value
	keyring.Delete(testKey, "test")
	return true
}

// IsKeyringAvailable returns whether the system keyring is available
func (c *Config) IsKeyringAvailable() bool {
	return c.keyringAvailable
}

// ConfigDir returns the config directory path
func (c *Config) ConfigDir() string {
	return c.configDir
}

// getOrCreateKey retrieves or generates the encryption key (for file fallback)
func (c *Config) getOrCreateKey() ([]byte, error) {
	keyPath := filepath.Join(c.configDir, keyFile)

	// Try to read existing key
	key, err := os.ReadFile(keyPath)
	if err == nil && len(key) == 32 {
		return key, nil
	}

	// Generate new key
	key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Save key with restrictive permissions
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save encryption key: %w", err)
	}

	return key, nil
}

// encrypt encrypts data using AES-GCM (for file fallback)
func (c *Config) encrypt(plaintext []byte) ([]byte, error) {
	key, err := c.getOrCreateKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM (for file fallback)
func (c *Config) decrypt(ciphertext []byte) ([]byte, error) {
	key, err := c.getOrCreateKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// saveTokensToKeyring saves tokens to the system keyring
func (c *Config) saveTokensToKeyring(tokens *model.TeslaTokens) error {
	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	if err := keyring.Set(keyringService, keyringUser, string(data)); err != nil {
		return fmt.Errorf("failed to save to keyring: %w", err)
	}

	return nil
}

// loadTokensFromKeyring loads tokens from the system keyring
func (c *Config) loadTokensFromKeyring() (*model.TeslaTokens, error) {
	data, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return nil, err
	}

	var tokens model.TeslaTokens
	if err := json.Unmarshal([]byte(data), &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	return &tokens, nil
}

// deleteTokensFromKeyring removes tokens from the system keyring
func (c *Config) deleteTokensFromKeyring() error {
	return keyring.Delete(keyringService, keyringUser)
}

// saveTokensToFile saves tokens to encrypted file (fallback)
func (c *Config) saveTokensToFile(tokens *model.TeslaTokens) error {
	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	encrypted, err := c.encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt tokens: %w", err)
	}

	// Encode as base64 for safe file storage
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	tokensPath := filepath.Join(c.configDir, tokensFile)
	if err := os.WriteFile(tokensPath, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("failed to write tokens file: %w", err)
	}

	return nil
}

// loadTokensFromFile loads tokens from encrypted file (fallback)
func (c *Config) loadTokensFromFile() (*model.TeslaTokens, error) {
	tokensPath := filepath.Join(c.configDir, tokensFile)

	encoded, err := os.ReadFile(tokensPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No tokens saved yet
		}
		return nil, fmt.Errorf("failed to read tokens file: %w", err)
	}

	encrypted, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		return nil, fmt.Errorf("failed to decode tokens: %w", err)
	}

	data, err := c.decrypt(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tokens: %w", err)
	}

	var tokens model.TeslaTokens
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tokens: %w", err)
	}

	return &tokens, nil
}

// deleteTokensFromFile removes tokens from encrypted file
func (c *Config) deleteTokensFromFile() error {
	tokensPath := filepath.Join(c.configDir, tokensFile)
	if err := os.Remove(tokensPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete tokens file: %w", err)
	}
	return nil
}

// SaveTokens saves tokens to secure storage (keyring with file fallback)
func (c *Config) SaveTokens(tokens *model.TeslaTokens) error {
	// Calculate expiry time if not set
	if tokens.ExpiresAt.IsZero() && tokens.ExpiresIn > 0 {
		tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	}

	// Try keyring first
	if c.keyringAvailable {
		if err := c.saveTokensToKeyring(tokens); err == nil {
			// Also delete any file-based tokens to avoid confusion
			c.deleteTokensFromFile()
			return nil
		}
		// Keyring failed, fall through to file
	}

	// Fall back to file-based storage
	return c.saveTokensToFile(tokens)
}

// LoadTokens loads tokens from secure storage (keyring with file fallback)
func (c *Config) LoadTokens() (*model.TeslaTokens, error) {
	// Try keyring first
	if c.keyringAvailable {
		tokens, err := c.loadTokensFromKeyring()
		if err == nil && tokens != nil {
			return tokens, nil
		}
		// Keyring failed or empty, try file
	}

	// Try file-based storage
	tokens, err := c.loadTokensFromFile()
	if err != nil {
		return nil, err
	}

	// If we loaded from file and keyring is available, migrate to keyring
	if tokens != nil && c.keyringAvailable {
		if err := c.saveTokensToKeyring(tokens); err == nil {
			// Migration successful, remove file
			c.deleteTokensFromFile()
		}
	}

	return tokens, nil
}

// DeleteTokens removes saved tokens (logout) from all storage
func (c *Config) DeleteTokens() error {
	var lastErr error

	// Delete from keyring
	if c.keyringAvailable {
		if err := c.deleteTokensFromKeyring(); err != nil {
			lastErr = err
		}
	}

	// Delete from file
	if err := c.deleteTokensFromFile(); err != nil {
		lastErr = err
	}

	return lastErr
}

// HasTokens checks if tokens are saved in any storage
func (c *Config) HasTokens() bool {
	// Check keyring
	if c.keyringAvailable {
		if _, err := c.loadTokensFromKeyring(); err == nil {
			return true
		}
	}

	// Check file
	tokensPath := filepath.Join(c.configDir, tokensFile)
	_, err := os.Stat(tokensPath)
	return err == nil
}

// GenerateCodeVerifier generates a PKCE code verifier
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCodeChallenge generates a PKCE code challenge from verifier
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateState generates a random state parameter
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
