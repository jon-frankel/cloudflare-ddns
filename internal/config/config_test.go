package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "config.toml")

	// Override the config path for this test
	oldPath := configPath
	configPath = testConfigPath
	defer func() { configPath = oldPath }()

	// Test saving
	cfg := Config{Hostname: "example.com"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Test loading
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Hostname != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", loaded.Hostname)
	}
}

func TestLoadNonexistent(t *testing.T) {
	// Override config path to a file that doesn't exist
	oldPath := configPath
	configPath = filepath.Join(t.TempDir(), "nonexistent.toml")
	defer func() { configPath = oldPath }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not error on missing file: %v", err)
	}

	if cfg.Hostname != "" {
		t.Errorf("Expected empty config, got %v", cfg)
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "config.toml")

	oldPath := configPath
	configPath = testConfigPath
	defer func() { configPath = oldPath }()

	if Exists() {
		t.Error("Expected Exists() to return false for nonexistent config")
	}

	// Create the config file
	cfg := Config{Hostname: "test.example.com"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if !Exists() {
		t.Error("Expected Exists() to return true after saving config")
	}
}

func TestMkdirOnSave(t *testing.T) {
	tmpDir := t.TempDir()
	testConfigPath := filepath.Join(tmpDir, "subdir", "config.toml")

	oldPath := configPath
	configPath = testConfigPath
	defer func() { configPath = oldPath }()

	cfg := Config{Hostname: "test.example.com"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify the directory was created
	if _, err := os.Stat(filepath.Dir(testConfigPath)); err != nil {
		t.Fatalf("Expected directory to be created: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testConfigPath); err != nil {
		t.Fatalf("Expected file to exist: %v", err)
	}
}
