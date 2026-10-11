// Package config loads and validates Ivo's startup configuration.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the provider credentials, endpoint, and model for a session.
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Load reads an optional .env file in the current directory, then reads the
// environment. Existing environment variables take precedence over the file.
// Loading .env populates the process environment; call Load once at startup.
// All three settings are required; no provider or model defaults are assumed.
func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		APIKey:  os.Getenv("API_KEY"),
		BaseURL: os.Getenv("BASE_URL"),
		Model:   os.Getenv("MODEL"),
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{"API_KEY", cfg.APIKey},
		{"BASE_URL", cfg.BaseURL},
		{"MODEL", cfg.Model},
	} {
		if strings.TrimSpace(field.value) == "" {
			return Config{}, fmt.Errorf("%s is required", field.name)
		}
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return Config{}, fmt.Errorf("BASE_URL must be an absolute HTTP(S) URL")
	}
	return cfg, nil
}
