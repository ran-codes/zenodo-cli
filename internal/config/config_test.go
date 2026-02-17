package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.DefaultProfile(); got != DefaultProfileName {
		t.Errorf("DefaultProfile() = %q, want %q", got, DefaultProfileName)
	}

	if got := cfg.ProfileBaseURL("production"); got != DefaultBaseURL {
		t.Errorf("ProfileBaseURL(production) = %q, want %q", got, DefaultBaseURL)
	}

	if got := cfg.ProfileBaseURL("sandbox"); got != SandboxBaseURL {
		t.Errorf("ProfileBaseURL(sandbox) = %q, want %q", got, SandboxBaseURL)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	cfg.SetDefaultProfile("sandbox")
	cfg.SetProfileValue("work", "base_url", "https://example.com/api")

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists with correct permissions.
	configPath := filepath.Join(tmpDir, appName, configFile)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("config file not found: %v", err)
	}
	// On Windows permissions work differently, skip perm check there.
	if perm := info.Mode().Perm(); perm&0077 != 0 && os.Getenv("OS") != "Windows_NT" {
		t.Errorf("config file permissions = %o, want 0600", perm)
	}

	// Reload and verify.
	cfg2, err := Load()
	if err != nil {
		t.Fatalf("Load() after save error: %v", err)
	}

	if got := cfg2.DefaultProfile(); got != "sandbox" {
		t.Errorf("after reload: DefaultProfile() = %q, want %q", got, "sandbox")
	}

	if got := cfg2.ProfileBaseURL("work"); got != "https://example.com/api" {
		t.Errorf("after reload: ProfileBaseURL(work) = %q, want %q", got, "https://example.com/api")
	}
}

func TestResolveBaseURL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name    string
		profile string
		sandbox bool
		want    string
	}{
		{"sandbox flag wins", "production", true, SandboxBaseURL},
		{"explicit profile", "sandbox", false, SandboxBaseURL},
		{"default profile", "", false, DefaultBaseURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.ResolveBaseURL(tt.profile, tt.sandbox)
			if got != tt.want {
				t.Errorf("ResolveBaseURL(%q, %v) = %q, want %q", tt.profile, tt.sandbox, got, tt.want)
			}
		})
	}
}

func TestResolveToken(t *testing.T) {
	orig := os.Getenv("ZENODO_TOKEN")
	defer os.Setenv("ZENODO_TOKEN", orig)

	os.Setenv("ZENODO_TOKEN", "env-token")

	if got := ResolveToken("explicit"); got != "explicit" {
		t.Errorf("explicit token: got %q, want %q", got, "explicit")
	}

	if got := ResolveToken(""); got != "env-token" {
		t.Errorf("env token: got %q, want %q", got, "env-token")
	}

	os.Unsetenv("ZENODO_TOKEN")
	if got := ResolveToken(""); got != "" {
		t.Errorf("no token: got %q, want empty", got)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abcdefghij", "****ghij"},
		{"ab", "****"},
		{"", "****"},
	}
	for _, tt := range tests {
		if got := MaskToken(tt.input); got != tt.want {
			t.Errorf("MaskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveProfile(t *testing.T) {
	orig := os.Getenv("ZENODO_PROFILE")
	defer os.Setenv("ZENODO_PROFILE", orig)

	if got := ResolveProfile("explicit"); got != "explicit" {
		t.Errorf("explicit: got %q, want %q", got, "explicit")
	}

	os.Setenv("ZENODO_PROFILE", "from-env")
	if got := ResolveProfile(""); got != "from-env" {
		t.Errorf("env: got %q, want %q", got, "from-env")
	}

	os.Unsetenv("ZENODO_PROFILE")
	if got := ResolveProfile(""); got != "" {
		t.Errorf("empty: got %q, want empty", got)
	}
}

func TestProfileNames(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	names := cfg.ProfileNames()
	if len(names) < 2 {
		t.Errorf("expected at least 2 default profiles, got %d", len(names))
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["production"] || !found["sandbox"] {
		t.Errorf("expected production and sandbox profiles, got %v", names)
	}
}
