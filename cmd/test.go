package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jonathan/cloudflare-ddns/internal/config"
	"github.com/jonathan/cloudflare-ddns/internal/keychain"
	"github.com/jonathan/cloudflare-ddns/internal/logger"
	"github.com/jonathan/cloudflare-ddns/internal/updater"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test the configuration and show one update cycle result",
	RunE:  doTest,
}

func doTest(cmd *cobra.Command, args []string) error {
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
		return err
	}

	if cfg.Hostname == "" {
		fmt.Fprintf(os.Stderr, "❌ Hostname not configured; run 'cloudflare-ddns' to complete setup\n")
		return fmt.Errorf("hostname not configured")
	}

	// Check keychain
	_, err = keychain.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ API key not configured in keychain: %v\n", err)
		return err
	}

	fmt.Printf("Testing configuration for: %s\n", cfg.Hostname)
	fmt.Println()

	// Run update
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := updater.RunOnce(ctx, cfg.Hostname)

	fmt.Printf("Hostname:       %s\n", cfg.Hostname)
	if result.CurrentIP != nil {
		fmt.Printf("Current IP:     %s\n", result.CurrentIP.String())
	}
	if result.RecordIP != nil {
		fmt.Printf("DNS Record IP:  %s\n", result.RecordIP.String())
	}
	fmt.Println()

	if result.Error != nil {
		fmt.Printf("❌ Test failed: %v\n", result.Error)
		return result.Error
	}

	if result.Updated {
		fmt.Printf("✓ DNS record updated successfully!\n")
	} else {
		fmt.Printf("✓ DNS record is already up to date\n")
	}

	return nil
}
