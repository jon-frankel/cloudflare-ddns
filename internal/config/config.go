package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Hostname string `toml:"hostname"`
}

var configPath string

func init() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	configPath = filepath.Join(configDir, "cloudflare-ddns", "config.toml")
}

// Load reads the configuration from disk. Returns empty config and nil if file doesn't exist.
func Load() (Config, error) {
	// Allow overriding hostname via environment variable (useful for Docker)
	if envHostname := os.Getenv("CLOUDFLARE_DDNS_HOSTNAME"); envHostname != "" {
		return Config{Hostname: envHostname}, nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}
	return cfg, nil
}

// Save writes the configuration to disk.
func Save(cfg Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// Exists returns true if config file exists.
func Exists() bool {
	_, err := os.Stat(configPath)
	return err == nil
}

// GetPath returns the config file path.
func GetPath() string {
	return configPath
}
