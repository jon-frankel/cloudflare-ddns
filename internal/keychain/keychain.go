package keychain

import (
	"errors"
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	service = "cloudflare-ddns"
	user    = "api-token"
)

var ErrNotConfigured = errors.New("keychain entry not configured")

// Set stores the API token in the system keychain.
func Set(token string) error {
	if err := keyring.Set(service, user, token); err != nil {
		return fmt.Errorf("failed to store token in keychain: %w", err)
	}
	return nil
}

// Get retrieves the API token from the system keychain.
func Get() (string, error) {
	// Allow overriding token via environment variable (useful for Docker)
	if envToken := os.Getenv("CLOUDFLARE_API_TOKEN"); envToken != "" {
		return envToken, nil
	}

	token, err := keyring.Get(service, user)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotConfigured
		}
		return "", fmt.Errorf("failed to retrieve token from keychain: %w", err)
	}
	return token, nil
}
