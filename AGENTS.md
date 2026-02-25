# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Go-based Dynamic DNS client that periodically updates Cloudflare DNS records with the current public IP address. Useful for servers with dynamic IPs that need consistent DNS names.

## Architecture

The application follows a typical Go CLI structure:
- **Main entry point**: Initializes configuration and starts the update loop
- **Cloudflare client**: Wraps the Cloudflare API to fetch zones, find records, and update DNS entries
- **IP detection**: Queries external services to determine the current public IP address
- **Configuration**: Likely loaded from environment variables or a config file (API token, domain name, record name, check interval)
- **Update loop**: Periodically checks IP and updates Cloudflare DNS if changed

Key architectural consideration: The updater should gracefully handle transient API failures and avoid thrashing the Cloudflare API on repeated identical IPs.

## Development Commands

Once `go.mod` is initialized:

```bash
# Build the binary
go build -o cloudflare-ddns

# Run tests (unit and integration where applicable)
go test ./...

# Run a specific test
go test -run TestName ./package

# Run with verbose output
go test -v ./...

# Check for linting issues (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...

# Run the application locally
go run main.go

# Show test coverage
go test -cover ./...
```

## Key Dependencies to Consider

- **Cloudflare API**: Use the official `github.com/cloudflare/cloudflare-go` SDK for API interactions
- **IP detection**: Consider using a public IP service API (e.g., ipify, checkip.amazonaws.com)
- **Configuration**: `github.com/caarlos0/env` for environment variable parsing or `github.com/spf13/viper` for config files

## Testing Strategy

- Unit tests for IP parsing and DNS record matching logic
- Mocked Cloudflare API calls for predictable testing
- Integration tests should use a test zone or skip if Cloudflare credentials aren't available

## Notes

- Secure credential handling: API tokens should only come from environment variables (never hardcoded)
- Consider idempotency: If the DNS record already has the correct IP, no update should be sent
- Error handling should distinguish between transient failures (retry) vs. configuration errors (exit)
- The project currently has no source code; start with `main.go` and establish the module structure early