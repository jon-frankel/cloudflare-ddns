package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/jonathan/cloudflare-ddns/internal/config"
	"github.com/jonathan/cloudflare-ddns/internal/keychain"
	"github.com/jonathan/cloudflare-ddns/internal/logger"
	"github.com/jonathan/cloudflare-ddns/internal/updater"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the DDNS update loop (runs every 60 seconds)",
	RunE:  doRun,
}

func doRun(cmd *cobra.Command, args []string) error {
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Hostname == "" {
		return fmt.Errorf("hostname not configured; run 'cloudflare-ddns' to complete setup")
	}

	// Check keychain
	_, err = keychain.Get()
	if err != nil {
		return fmt.Errorf("API key not configured in keychain: %w", err)
	}

	fmt.Printf("Starting DDNS update loop for %s\n", cfg.Hostname)
	slog.Info("Starting DDNS update loop", "hostname", cfg.Hostname)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run first update immediately
	result := updater.RunOnce(ctx, cfg.Hostname)
	logUpdateResult(cfg.Hostname, result)

	// Start the update loop
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result := updater.RunOnce(ctx, cfg.Hostname)
			logUpdateResult(cfg.Hostname, result)

		case <-sigChan:
			fmt.Println("\nShutting down...")
			slog.Info("Shutting down DDNS update loop")
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func logUpdateResult(hostname string, result updater.UpdateResult) {
	if result.Error != nil {
		slog.Error("Update cycle failed", "hostname", hostname, "error", result.Error)
		fmt.Printf("❌ Update failed: %v\n", result.Error)
		return
	}

	if result.Updated {
		fmt.Printf("✓ DNS record updated: %s -> %s\n", result.RecordIP, result.CurrentIP)
	} else {
		fmt.Printf("ℹ DNS record is current: %s\n", result.CurrentIP)
	}
}
