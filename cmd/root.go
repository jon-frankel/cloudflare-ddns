package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jonathan/cloudflare-ddns/internal/config"
	"github.com/jonathan/cloudflare-ddns/internal/keychain"
	"github.com/jonathan/cloudflare-ddns/internal/logger"
	"github.com/jonathan/cloudflare-ddns/internal/updater"
)

var (
	version = "dev"
	rootCmd = &cobra.Command{
		Use:     "cloudflare-ddns",
		Short:   "Dynamic DNS client for Cloudflare",
		Long:    "cloudflare-ddns keeps a Cloudflare DNS A record in sync with your machine's public IP address",
		Version: version,
		RunE:    runRoot,
	}
)

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(logsCmd)
}

func runRoot(cmd *cobra.Command, args []string) error {
	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
	}

	// Check if config and keychain exist
	_, err := keychain.Get()
	if !config.Exists() || err != nil {
		// Run setup
		return setupFlow()
	}

	fmt.Println("Configuration found. Use 'cloudflare-ddns run' to start the daemon or 'cloudflare-ddns test' to verify.")
	return nil
}

func setupFlow() error {
	fmt.Println("Welcome to cloudflare-ddns! Let's set up your configuration.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Prompt for hostname
	fmt.Print("Hostname to update (e.g. home.example.com): ")
	hostname, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read hostname: %w", err)
	}
	hostname = strings.TrimSpace(hostname)

	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	// Prompt for API key
	fmt.Print("Cloudflare API key: ")
	keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}
	fmt.Println() // newline after password input
	apiKey := strings.TrimSpace(string(keyBytes))

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Test credentials
	fmt.Println("Testing credentials and setting up DNS record...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Temporarily set the keychain entry for testing
	if err := keychain.Set(apiKey); err != nil {
		return fmt.Errorf("failed to store API key: %w", err)
	}

	// Use RunOnceWithCreate to create the record if it doesn't exist
	result := updater.RunOnceWithCreate(ctx, hostname)
	if result.Error != nil {
		fmt.Printf("❌ Setup failed: %v\n", result.Error)
		return result.Error
	}

	fmt.Println("✓ Credentials validated successfully!")
	fmt.Printf("  Current IP: %s\n", result.CurrentIP.String())
	fmt.Printf("  DNS Record IP: %s\n", result.RecordIP.String())
	if result.Updated {
		fmt.Println("  (DNS record was created)")
	} else {
		fmt.Println("  (DNS record was already up to date)")
	}
	fmt.Println()

	// Save configuration
	cfg := config.Config{Hostname: hostname}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Configuration saved to %s\n", config.GetPath())
	fmt.Println()
	fmt.Println("Setup complete! You can now run:")
	fmt.Println("  cloudflare-ddns run     - Start the daemon (runs every 60 seconds)")
	fmt.Println("  cloudflare-ddns test    - Test the connection and show current status")
	fmt.Println("  cloudflare-ddns logs    - View recent log entries")
	fmt.Println()
	fmt.Println("To run as a background service on macOS:")
	fmt.Println("  brew services start cloudflare-ddns")

	return nil
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets the application version.
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}
