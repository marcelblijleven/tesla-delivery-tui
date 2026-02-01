package config

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/model"
)

func TestNew(t *testing.T) {
	cfg, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("New() returned nil")
	}

	if cfg.configDir == "" {
		t.Error("configDir is empty")
	}

	// Check that config directory exists
	if _, err := os.Stat(cfg.configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %s", cfg.configDir)
	}
}

func TestConfig_ConfigDir(t *testing.T) {
	cfg, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	dir := cfg.ConfigDir()
	if dir == "" {
		t.Error("ConfigDir() returned empty string")
	}

	// Should contain our app name
	if filepath.Base(dir) != appName {
		t.Errorf("ConfigDir() = %s, want to end with %s", dir, appName)
	}
}

func TestConfig_EncryptDecrypt(t *testing.T) {
	// Create a temp directory for this test
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir}

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello world"},
		{"json data", `{"access_token":"abc123","refresh_token":"def456"}`},
		{"empty string", ""},
		{"unicode", "Hello 世界 🌍"},
		{"long text", string(make([]byte, 10000))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := cfg.encrypt([]byte(tc.plaintext))
			if err != nil {
				t.Fatalf("encrypt() error = %v", err)
			}

			// Encrypted should be different from plaintext (unless empty)
			if tc.plaintext != "" && string(encrypted) == tc.plaintext {
				t.Error("Encrypted data equals plaintext")
			}

			decrypted, err := cfg.decrypt(encrypted)
			if err != nil {
				t.Fatalf("decrypt() error = %v", err)
			}

			if string(decrypted) != tc.plaintext {
				t.Errorf("decrypt() = %q, want %q", string(decrypted), tc.plaintext)
			}
		})
	}
}

func TestConfig_EncryptDecrypt_DifferentCiphertexts(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir}

	plaintext := []byte("test data")

	// Encrypt the same plaintext twice
	encrypted1, err := cfg.encrypt(plaintext)
	if err != nil {
		t.Fatalf("First encrypt() error = %v", err)
	}

	encrypted2, err := cfg.encrypt(plaintext)
	if err != nil {
		t.Fatalf("Second encrypt() error = %v", err)
	}

	// Due to random nonce, ciphertexts should be different
	if string(encrypted1) == string(encrypted2) {
		t.Error("Two encryptions of same plaintext produced identical ciphertexts (nonce should make them different)")
	}

	// But both should decrypt to the same plaintext
	decrypted1, _ := cfg.decrypt(encrypted1)
	decrypted2, _ := cfg.decrypt(encrypted2)

	if string(decrypted1) != string(decrypted2) {
		t.Error("Decrypted values don't match")
	}
}

func TestConfig_Decrypt_InvalidData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir}

	// Need to create a key first
	_, _ = cfg.encrypt([]byte("init"))

	testCases := []struct {
		name string
		data []byte
	}{
		{"empty data", []byte{}},
		{"too short", []byte("abc")},
		{"random garbage", []byte("this is not encrypted data at all")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := cfg.decrypt(tc.data)
			if err == nil {
				t.Error("decrypt() should have returned an error for invalid data")
			}
		})
	}
}

func TestConfig_SaveLoadTokens_File(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config without keyring
	cfg := &Config{configDir: tempDir, keyringAvailable: false}

	tokens := &model.TeslaTokens{
		AccessToken:  "access123",
		RefreshToken: "refresh456",
		ExpiresIn:    3600,
		Scope:        "openid email",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	// Save tokens
	if err := cfg.SaveTokens(tokens); err != nil {
		t.Fatalf("SaveTokens() error = %v", err)
	}

	// Load tokens
	loaded, err := cfg.LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("LoadTokens() returned nil")
	}

	if loaded.AccessToken != tokens.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, tokens.AccessToken)
	}
	if loaded.RefreshToken != tokens.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, tokens.RefreshToken)
	}
}

func TestConfig_SaveTokens_SetsExpiresAt(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir, keyringAvailable: false}

	tokens := &model.TeslaTokens{
		AccessToken:  "access123",
		RefreshToken: "refresh456",
		ExpiresIn:    3600, // 1 hour
		// ExpiresAt is zero
	}

	before := time.Now()
	if err := cfg.SaveTokens(tokens); err != nil {
		t.Fatalf("SaveTokens() error = %v", err)
	}
	after := time.Now()

	// ExpiresAt should now be set
	if tokens.ExpiresAt.IsZero() {
		t.Error("ExpiresAt was not set")
	}

	// Should be approximately 1 hour from now
	expectedMin := before.Add(time.Duration(tokens.ExpiresIn) * time.Second)
	expectedMax := after.Add(time.Duration(tokens.ExpiresIn) * time.Second)

	if tokens.ExpiresAt.Before(expectedMin) || tokens.ExpiresAt.After(expectedMax) {
		t.Errorf("ExpiresAt = %v, want between %v and %v", tokens.ExpiresAt, expectedMin, expectedMax)
	}
}

func TestConfig_LoadTokens_NoFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir, keyringAvailable: false}

	// Load without saving first
	tokens, err := cfg.LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens() error = %v", err)
	}

	if tokens != nil {
		t.Errorf("LoadTokens() = %v, want nil", tokens)
	}
}

func TestConfig_DeleteTokens(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir, keyringAvailable: false}

	tokens := &model.TeslaTokens{
		AccessToken:  "access123",
		RefreshToken: "refresh456",
	}

	// Save then delete
	if err := cfg.SaveTokens(tokens); err != nil {
		t.Fatalf("SaveTokens() error = %v", err)
	}

	if err := cfg.DeleteTokens(); err != nil {
		t.Fatalf("DeleteTokens() error = %v", err)
	}

	// Load should return nil
	loaded, err := cfg.LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens() error = %v", err)
	}

	if loaded != nil {
		t.Error("LoadTokens() should return nil after delete")
	}
}

func TestConfig_HasTokens(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir, keyringAvailable: false}

	// Initially no tokens
	if cfg.HasTokens() {
		t.Error("HasTokens() = true, want false (no tokens saved)")
	}

	// Save tokens
	tokens := &model.TeslaTokens{AccessToken: "test"}
	if err := cfg.SaveTokens(tokens); err != nil {
		t.Fatalf("SaveTokens() error = %v", err)
	}

	// Now should have tokens
	if !cfg.HasTokens() {
		t.Error("HasTokens() = false, want true (tokens saved)")
	}

	// Delete tokens
	if err := cfg.DeleteTokens(); err != nil {
		t.Fatalf("DeleteTokens() error = %v", err)
	}

	// Should not have tokens again
	if cfg.HasTokens() {
		t.Error("HasTokens() = true, want false (tokens deleted)")
	}
}

func TestConfig_TokenFilePermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir, keyringAvailable: false}

	tokens := &model.TeslaTokens{AccessToken: "test"}
	if err := cfg.SaveTokens(tokens); err != nil {
		t.Fatalf("SaveTokens() error = %v", err)
	}

	tokensPath := filepath.Join(tempDir, tokensFile)
	info, err := os.Stat(tokensPath)
	if err != nil {
		t.Fatalf("Failed to stat tokens file: %v", err)
	}

	// Check file permissions are restricted (0600)
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Token file permissions = %o, want 0600", mode)
	}
}

func TestGenerateCodeVerifier(t *testing.T) {
	verifier1, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("GenerateCodeVerifier() error = %v", err)
	}

	if verifier1 == "" {
		t.Error("GenerateCodeVerifier() returned empty string")
	}

	// Should be base64url encoded 32 bytes = 43 characters
	if len(verifier1) != 43 {
		t.Errorf("GenerateCodeVerifier() length = %d, want 43", len(verifier1))
	}

	// Should be valid base64url
	_, err = base64.RawURLEncoding.DecodeString(verifier1)
	if err != nil {
		t.Errorf("GenerateCodeVerifier() is not valid base64url: %v", err)
	}

	// Generate another one - should be different
	verifier2, _ := GenerateCodeVerifier()
	if verifier1 == verifier2 {
		t.Error("Two calls to GenerateCodeVerifier() returned same value")
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

	challenge := GenerateCodeChallenge(verifier)
	if challenge == "" {
		t.Error("GenerateCodeChallenge() returned empty string")
	}

	// Should be base64url encoded SHA256 = 43 characters
	if len(challenge) != 43 {
		t.Errorf("GenerateCodeChallenge() length = %d, want 43", len(challenge))
	}

	// Should be valid base64url
	_, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil {
		t.Errorf("GenerateCodeChallenge() is not valid base64url: %v", err)
	}

	// Same verifier should always produce same challenge (deterministic)
	challenge2 := GenerateCodeChallenge(verifier)
	if challenge != challenge2 {
		t.Error("GenerateCodeChallenge() not deterministic")
	}

	// Different verifier should produce different challenge
	challenge3 := GenerateCodeChallenge("different-verifier")
	if challenge == challenge3 {
		t.Error("Different verifiers produced same challenge")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error = %v", err)
	}

	if state1 == "" {
		t.Error("GenerateState() returned empty string")
	}

	// Should be base64url encoded 16 bytes = 22 characters
	if len(state1) != 22 {
		t.Errorf("GenerateState() length = %d, want 22", len(state1))
	}

	// Should be valid base64url
	_, err = base64.RawURLEncoding.DecodeString(state1)
	if err != nil {
		t.Errorf("GenerateState() is not valid base64url: %v", err)
	}

	// Generate another one - should be different
	state2, _ := GenerateState()
	if state1 == state2 {
		t.Error("Two calls to GenerateState() returned same value")
	}
}

func TestConfig_getOrCreateKey(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tesla-delivery-tui-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &Config{configDir: tempDir}

	// First call should create key
	key1, err := cfg.getOrCreateKey()
	if err != nil {
		t.Fatalf("getOrCreateKey() error = %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("Key length = %d, want 32", len(key1))
	}

	// Second call should return same key
	key2, err := cfg.getOrCreateKey()
	if err != nil {
		t.Fatalf("getOrCreateKey() second call error = %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("getOrCreateKey() returned different keys on second call")
	}

	// Check key file permissions
	keyPath := filepath.Join(tempDir, keyFile)
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Key file permissions = %o, want 0600", mode)
	}
}

func TestConfig_IsKeyringAvailable(t *testing.T) {
	cfg := &Config{keyringAvailable: true}
	if !cfg.IsKeyringAvailable() {
		t.Error("IsKeyringAvailable() = false, want true")
	}

	cfg = &Config{keyringAvailable: false}
	if cfg.IsKeyringAvailable() {
		t.Error("IsKeyringAvailable() = true, want false")
	}
}
