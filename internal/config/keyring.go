package config

import (
	"fmt"
	"log/slog"

	"github.com/zalando/go-keyring"
)

const keyringService = "zenodo-cli"

// KeyringProvider abstracts keyring operations for testing.
type KeyringProvider interface {
	Get(service, user string) (string, error)
	Set(service, user, password string) error
	Delete(service, user string) error
}

// OSKeyring uses the real OS keyring via go-keyring.
type OSKeyring struct{}

func (OSKeyring) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (OSKeyring) Set(service, user, password string) error {
	return keyring.Set(service, user, password)
}

func (OSKeyring) Delete(service, user string) error {
	return keyring.Delete(service, user)
}

// Keyring manages token storage in the OS keyring with config file fallback.
type Keyring struct {
	provider KeyringProvider
	available bool
}

// NewKeyring creates a Keyring using the real OS keyring.
func NewKeyring() *Keyring {
	return NewKeyringWith(OSKeyring{})
}

// NewKeyringWith creates a Keyring with a custom provider (for testing).
func NewKeyringWith(provider KeyringProvider) *Keyring {
	kr := &Keyring{provider: provider}
	kr.available = kr.probe()
	return kr
}

// probe checks if the keyring is functional by doing a test set/get/delete cycle.
func (k *Keyring) probe() bool {
	const testUser = "__zenodo_cli_probe__"
	if err := k.provider.Set(keyringService, testUser, "probe"); err != nil {
		return false
	}
	k.provider.Delete(keyringService, testUser)
	return true
}

// Available returns whether the OS keyring is functional.
func (k *Keyring) Available() bool {
	return k.available
}

// keyringUser returns the keyring entry name for a profile.
// Tokens are stored as "zenodo-cli:<profile>" in the OS keyring.
func keyringUser(profile string) string {
	return profile
}

// GetToken retrieves the token for a profile from the keyring.
// Returns empty string if not found or keyring unavailable.
func (k *Keyring) GetToken(profile string) string {
	if !k.available {
		return ""
	}
	token, err := k.provider.Get(keyringService, keyringUser(profile))
	if err != nil {
		return ""
	}
	return token
}

// SetToken stores a token for a profile in the keyring.
func (k *Keyring) SetToken(profile, token string) error {
	if !k.available {
		return fmt.Errorf("keyring not available; token will be stored in config file instead")
	}
	return k.provider.Set(keyringService, keyringUser(profile), token)
}

// DeleteToken removes a token for a profile from the keyring.
func (k *Keyring) DeleteToken(profile string) error {
	if !k.available {
		return nil
	}
	return k.provider.Delete(keyringService, keyringUser(profile))
}

// MigrateToken checks if a plaintext token exists in the config for the given
// profile and migrates it to the keyring, removing it from the config file.
func (k *Keyring) MigrateToken(cfg *Config, profile string) {
	if !k.available {
		return
	}

	token := cfg.GetProfileValue(profile, "token")
	if token == "" {
		return
	}

	if err := k.SetToken(profile, token); err != nil {
		slog.Warn("failed to migrate token to keyring", "profile", profile, "error", err)
		return
	}

	// Remove plaintext token from config.
	cfg.SetProfileValue(profile, "token", "")
	if err := cfg.Save(); err != nil {
		slog.Warn("failed to save config after token migration", "error", err)
		return
	}

	slog.Info("migrated token from config file to keyring", "profile", profile)
}

// ResolveTokenFull returns the token from the highest-priority source.
// Priority: explicit flag > ZENODO_TOKEN env var > keyring > config file.
func ResolveTokenFull(explicit string, kr *Keyring, cfg *Config, profile string) string {
	// 1. Explicit flag
	if token := ResolveToken(explicit); token != "" {
		return token
	}

	// 2. Keyring
	if kr != nil {
		if token := kr.GetToken(profile); token != "" {
			return token
		}
	}

	// 3. Config file fallback
	if cfg != nil {
		if token := cfg.GetProfileValue(profile, "token"); token != "" {
			return token
		}
	}

	return ""
}
