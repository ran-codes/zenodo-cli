package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigDir_XDGOverride(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", orig)

	os.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")

	dir := GetConfigDir()
	expected := filepath.Join("/tmp/test-xdg", appName)
	if dir != expected {
		t.Errorf("got %q, want %q", dir, expected)
	}
}

func TestGetConfigDir_Default(t *testing.T) {
	orig := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", orig)

	os.Unsetenv("XDG_CONFIG_HOME")

	dir := GetConfigDir()
	if !strings.Contains(dir, appName) {
		t.Errorf("expected config dir to contain %q, got %q", appName, dir)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	path := GetConfigFilePath()
	if !strings.HasSuffix(path, configFile) {
		t.Errorf("expected path to end with %q, got %q", configFile, path)
	}
}
