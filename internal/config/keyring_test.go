package config

import (
	"fmt"
	"os"
	"testing"
)

// MockKeyring is an in-memory keyring for testing.
type MockKeyring struct {
	store map[string]string
}

func NewMockKeyring() *MockKeyring {
	return &MockKeyring{store: make(map[string]string)}
}

func (m *MockKeyring) Get(service, user string) (string, error) {
	key := service + ":" + user
	if val, ok := m.store[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("not found")
}

func (m *MockKeyring) Set(service, user, password string) error {
	key := service + ":" + user
	m.store[key] = password
	return nil
}

func (m *MockKeyring) Delete(service, user string) error {
	key := service + ":" + user
	delete(m.store, key)
	return nil
}

// FailingKeyring simulates an unavailable keyring.
type FailingKeyring struct{}

func (FailingKeyring) Get(service, user string) (string, error) {
	return "", fmt.Errorf("keyring unavailable")
}

func (FailingKeyring) Set(service, user, password string) error {
	return fmt.Errorf("keyring unavailable")
}

func (FailingKeyring) Delete(service, user string) error {
	return fmt.Errorf("keyring unavailable")
}

func TestKeyringSetGetDelete(t *testing.T) {
	kr := NewKeyringWith(NewMockKeyring())

	if !kr.Available() {
		t.Fatal("mock keyring should be available")
	}

	if err := kr.SetToken("production", "my-secret-token"); err != nil {
		t.Fatalf("SetToken error: %v", err)
	}

	got := kr.GetToken("production")
	if got != "my-secret-token" {
		t.Errorf("GetToken() = %q, want %q", got, "my-secret-token")
	}

	if err := kr.DeleteToken("production"); err != nil {
		t.Fatalf("DeleteToken error: %v", err)
	}

	got = kr.GetToken("production")
	if got != "" {
		t.Errorf("GetToken() after delete = %q, want empty", got)
	}
}

func TestKeyringUnavailableFallback(t *testing.T) {
	kr := NewKeyringWith(FailingKeyring{})

	if kr.Available() {
		t.Fatal("failing keyring should not be available")
	}

	// SetToken should return error when keyring unavailable.
	err := kr.SetToken("production", "token")
	if err == nil {
		t.Error("expected error from SetToken with unavailable keyring")
	}

	// GetToken should return empty string.
	if got := kr.GetToken("production"); got != "" {
		t.Errorf("GetToken() with unavailable keyring = %q, want empty", got)
	}

	// DeleteToken should not error (no-op).
	if err := kr.DeleteToken("production"); err != nil {
		t.Errorf("DeleteToken() with unavailable keyring should not error: %v", err)
	}
}

func TestKeyringMultipleProfiles(t *testing.T) {
	kr := NewKeyringWith(NewMockKeyring())

	kr.SetToken("production", "prod-token")
	kr.SetToken("sandbox", "sandbox-token")

	if got := kr.GetToken("production"); got != "prod-token" {
		t.Errorf("production token = %q, want %q", got, "prod-token")
	}
	if got := kr.GetToken("sandbox"); got != "sandbox-token" {
		t.Errorf("sandbox token = %q, want %q", got, "sandbox-token")
	}
}

func TestMigrateToken(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Set up config with a plaintext token.
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	cfg.SetProfileValue("production", "token", "plaintext-secret")
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Migrate.
	kr := NewKeyringWith(NewMockKeyring())
	kr.MigrateToken(cfg, "production")

	// Token should now be in keyring.
	if got := kr.GetToken("production"); got != "plaintext-secret" {
		t.Errorf("keyring token after migration = %q, want %q", got, "plaintext-secret")
	}

	// Token should be removed from config.
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after migration error: %v", err)
	}
	if got := cfg2.GetProfileValue("production", "token"); got != "" {
		t.Errorf("config token after migration = %q, want empty", got)
	}
}

func TestResolveTokenFull(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	origEnv := os.Getenv("ZENODO_TOKEN")
	defer os.Setenv("ZENODO_TOKEN", origEnv)
	os.Unsetenv("ZENODO_TOKEN")

	cfg, _ := Load()
	cfg.SetProfileValue("production", "token", "config-token")
	cfg.Save()
	cfg, _ = Load()

	kr := NewKeyringWith(NewMockKeyring())
	kr.SetToken("production", "keyring-token")

	// Explicit flag wins.
	got := ResolveTokenFull("flag-token", kr, cfg, "production")
	if got != "flag-token" {
		t.Errorf("explicit flag: got %q, want %q", got, "flag-token")
	}

	// Env var wins over keyring.
	os.Setenv("ZENODO_TOKEN", "env-token")
	got = ResolveTokenFull("", kr, cfg, "production")
	if got != "env-token" {
		t.Errorf("env var: got %q, want %q", got, "env-token")
	}
	os.Unsetenv("ZENODO_TOKEN")

	// Keyring wins over config.
	got = ResolveTokenFull("", kr, cfg, "production")
	if got != "keyring-token" {
		t.Errorf("keyring: got %q, want %q", got, "keyring-token")
	}

	// Config fallback.
	kr2 := NewKeyringWith(NewMockKeyring()) // empty keyring
	got = ResolveTokenFull("", kr2, cfg, "production")
	if got != "config-token" {
		t.Errorf("config fallback: got %q, want %q", got, "config-token")
	}

	// Nothing available.
	cfg2, _ := Load()
	got = ResolveTokenFull("", kr2, cfg2, "nonexistent")
	if got != "" {
		t.Errorf("no token: got %q, want empty", got)
	}
}
