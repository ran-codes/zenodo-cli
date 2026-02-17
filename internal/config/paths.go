package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	appName    = "zenodo-cli"
	configFile = "config.yaml"
)

// GetConfigDir returns the configuration directory path.
// Uses $XDG_CONFIG_HOME/zenodo-cli on Linux/macOS, or %APPDATA%/zenodo-cli on Windows.
func GetConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}

	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, appName)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", appName)
	}
	return filepath.Join(home, ".config", appName)
}

// GetConfigFilePath returns the full path to the config file.
func GetConfigFilePath() string {
	return filepath.Join(GetConfigDir(), configFile)
}
