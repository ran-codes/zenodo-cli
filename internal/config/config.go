package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultBaseURL        = "https://zenodo.org/api"
	SandboxBaseURL        = "https://sandbox.zenodo.org/api"
	DefaultProfileName    = "production"
	configFilePermissions = 0600
	configDirPermissions  = 0700
)

// Config holds the application configuration backed by viper.
type Config struct {
	v          *viper.Viper
	deletedKeys []string
}

// Load reads the config file and returns a Config instance.
// If the config file doesn't exist, a default config is created.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	configDir := GetConfigDir()
	v.AddConfigPath(configDir)
	v.SetConfigName("config")

	// Defaults
	v.SetDefault("default_profile", DefaultProfileName)
	v.SetDefault("profiles.production.base_url", DefaultBaseURL)
	v.SetDefault("profiles.sandbox.base_url", SandboxBaseURL)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// Config file doesn't exist yet — that's fine, we'll use defaults.
	}

	return &Config{v: v}, nil
}

// Save writes the current configuration to disk.
func (c *Config) Save() error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, configDirPermissions); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	configPath := GetConfigFilePath()

	// Apply pending deletes before writing.
	if len(c.deletedKeys) > 0 {
		settings := c.v.AllSettings()
		for _, key := range c.deletedKeys {
			parts := strings.Split(key, ".")
			deleteNestedKey(settings, parts)
		}
		c.deletedKeys = nil
		// Write cleaned settings directly via a fresh viper instance.
		fresh := viper.New()
		fresh.SetConfigType("yaml")
		if err := fresh.MergeConfigMap(settings); err != nil {
			return fmt.Errorf("applying deletes: %w", err)
		}
		if err := fresh.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	} else {
		if err := c.v.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}

	// Enforce 0600 permissions (no-op on Windows).
	if runtime.GOOS != "windows" {
		if err := os.Chmod(configPath, configFilePermissions); err != nil {
			return fmt.Errorf("setting config permissions: %w", err)
		}
	}

	return nil
}

// DefaultProfile returns the name of the default profile.
func (c *Config) DefaultProfile() string {
	return c.v.GetString("default_profile")
}

// SetDefaultProfile sets the default profile name.
func (c *Config) SetDefaultProfile(name string) {
	c.v.Set("default_profile", name)
}

// ProfileBaseURL returns the base URL for the given profile.
func (c *Config) ProfileBaseURL(profile string) string {
	url := c.v.GetString(profileKey(profile, "base_url"))
	if url == "" {
		return DefaultBaseURL
	}
	return url
}

// SetProfileValue sets a key within a profile's config section.
func (c *Config) SetProfileValue(profile, key, value string) {
	c.v.Set(profileKey(profile, key), value)
}

// GetProfileValue reads a key from a profile's config section.
func (c *Config) GetProfileValue(profile, key string) string {
	return c.v.GetString(profileKey(profile, key))
}

// ProfileNames returns all configured profile names.
func (c *Config) ProfileNames() []string {
	profiles := c.v.GetStringMap("profiles")
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	return names
}

// ResolveBaseURL determines the base URL given flags and profile.
// Priority: --sandbox flag > profile's base_url > default.
func (c *Config) ResolveBaseURL(profile string, sandbox bool) string {
	if sandbox {
		return SandboxBaseURL
	}
	if profile == "" {
		profile = c.DefaultProfile()
	}
	return c.ProfileBaseURL(profile)
}

// Get returns a raw config value by dotted key path.
func (c *Config) Get(key string) interface{} {
	return c.v.Get(key)
}

// Set sets a raw config value by dotted key path.
func (c *Config) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

// Delete removes a key from the configuration.
// Viper doesn't support deleting keys natively, so we track them
// and remove them from the settings map during Save.
func (c *Config) Delete(key string) {
	c.deletedKeys = append(c.deletedKeys, strings.ToLower(key))
}

// deleteNestedKey removes a key from a nested map given a dotted key path split into parts.
func deleteNestedKey(m map[string]interface{}, parts []string) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		delete(m, parts[0])
		return
	}
	next, ok := m[parts[0]]
	if !ok {
		return
	}
	nextMap, ok := next.(map[string]interface{})
	if !ok {
		return
	}
	deleteNestedKey(nextMap, parts[1:])
}

func profileKey(profile, key string) string {
	return strings.Join([]string{"profiles", profile, key}, ".")
}

// ResolveProfile returns the effective profile name.
// Priority: explicit profile arg > ZENODO_PROFILE env > default_profile config.
func ResolveProfile(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv("ZENODO_PROFILE"); env != "" {
		return env
	}
	return ""
}

// ResolveToken returns the token from the highest-priority source.
// Priority: explicit flag > ZENODO_TOKEN env var.
// Keyring lookup is handled separately (Issue #3).
func ResolveToken(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv("ZENODO_TOKEN"); env != "" {
		return env
	}
	return ""
}

// MaskToken returns a masked version of a token for safe display.
func MaskToken(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return "****" + token[len(token)-4:]
}

// CheckConfigPermissions warns if the config file has overly broad permissions.
// Skipped on Windows where Unix-style file permissions don't apply.
func CheckConfigPermissions() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	path := GetConfigFilePath()
	info, err := os.Stat(path)
	if err != nil {
		return nil // File doesn't exist yet, no problem.
	}
	perm := info.Mode().Perm()
	if perm&0077 != 0 {
		return fmt.Errorf("config file %s has permissions %o (expected 0600); run: chmod 600 %s",
			filepath.Base(path), perm, path)
	}
	return nil
}
