package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jon-frankel/cloudflare-ddns/internal/config"
	"github.com/jon-frankel/cloudflare-ddns/internal/keychain"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "View and modify the application configuration.",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE:  runConfigShow,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key=value] [key]...",
	Short: "Update configuration",
	Long: `Update configuration values.
If no arguments are provided, runs the interactive setup wizard.
Supported keys: hostname, token

Examples:
  cloudflare-ddns config set hostname=new.example.com
  cloudflare-ddns config set token
  cloudflare-ddns config set hostname token`,
	RunE: runConfigSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigShow(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Hostname: %s\n", cfg.Hostname)

	token, err := keychain.Get()
	if err != nil {
		fmt.Printf("Token:    <not configured or error: %v>\n", err)
	} else {
		masked := maskToken(token)
		fmt.Printf("Token:    %s\n", masked)
	}
	return nil
}

func runConfigSet(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return setupFlow()
	}

	cfg, err := config.Load()
	if err != nil {
		// If config doesn't exist, we start with an empty one
		cfg = config.Config{}
	}

	var newToken string
	tokenUpdated := false
	configUpdated := false

	for _, arg := range args {
		key, value, hasValue := strings.Cut(arg, "=")
		key = strings.ToLower(key)

		switch key {
		case "hostname":
			if !hasValue {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Hostname: ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read hostname: %w", err)
				}
				value = strings.TrimSpace(input)
			}
			if value == "" {
				return fmt.Errorf("hostname cannot be empty")
			}
			cfg.Hostname = value
			configUpdated = true

		case "token", "api-token", "key":
			if !hasValue {
				fmt.Print("Cloudflare API key: ")
				bytePw, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to read API key: %w", err)
				}
				fmt.Println() // newline after password input
				value = strings.TrimSpace(string(bytePw))
			}
			if value == "" {
				return fmt.Errorf("token cannot be empty")
			}
			newToken = value
			tokenUpdated = true

		default:
			return fmt.Errorf("unknown configuration key: %s", key)
		}
	}

	if configUpdated {
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("Configuration saved.")
	}

	if tokenUpdated {
		if err := keychain.Set(newToken); err != nil {
			return fmt.Errorf("failed to save token to keychain: %w", err)
		}
		fmt.Println("Token saved to keychain.")
	}

	return nil
}

func maskToken(token string) string {
	if len(token) <= 4 {
		return "****"
	}
	return "****" + token[len(token)-4:]
}
