package logger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Init initializes the global logger with JSON output and log rotation.
func Init() error {
	logDir, err := getLogDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(logDir, 0700); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, "cloudflare-ddns.log")

	// Create lumberjack logger with rotation
	lj := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,   // MB
		MaxBackups: 3,
		Compress:   true,
	}

	// Create slog handler with JSON output
	handler := slog.NewJSONHandler(lj, nil)
	slog.SetDefault(slog.New(handler))

	return nil
}

// getLogDir returns the log directory based on OS.
func getLogDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback for macOS: ~/Library/Logs
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			return "", fmt.Errorf("unable to determine log directory")
		}
		return filepath.Join(homeDir, "Library", "Logs", "cloudflare-ddns"), nil
	}
	return filepath.Join(cacheDir, "cloudflare-ddns"), nil
}

// GetLogFilePath returns the full path to the log file.
func GetLogFilePath() (string, error) {
	logDir, err := getLogDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(logDir, "cloudflare-ddns.log"), nil
}
