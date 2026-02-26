# cloudflare-ddns

A Go-based Dynamic DNS client that keeps a Cloudflare DNS A record in sync with your machine's public IP address. Useful for servers with dynamic IPs that need consistent DNS names.

## Features

- **Automatic IP Updates**: Periodically checks your public IP and updates Cloudflare DNS records automatically
- **System Keychain Integration**: Securely stores API keys in system keychain (macOS, Linux, Windows)
- **First-Run Setup**: Interactive setup wizard guides you through configuration
- **Background Service**: Runs as a background daemon (60-second polling interval)
- **Graceful Error Handling**: Retries on transient failures, distinguishes configuration errors
- **JSON Logging**: Structured logs with automatic rotation
- **Homebrew Integration**: Easy installation and management via Homebrew on macOS/Linux
- **Docker Support**: Run as a lightweight container

## Installation

### From Source

```bash
git clone https://github.com/jon-frankel/cloudflare-ddns.git
cd cloudflare-ddns
go build -o cloudflare-ddns
```

### Via Homebrew

```bash
brew tap jon-frankel/cloudflare-ddns
brew install cloudflare-ddns
```

### Via Docker

You can run the application as a Docker container. This is useful for NAS devices or servers where you don't want to install Go or manage system services.

**Pull the image:**
```bash
docker pull ghcr.io/jon-frankel/cloudflare-ddns:latest
```

**Run the container:**
```bash
docker run -d \
  --name cloudflare-ddns \
  --restart unless-stopped \
  -e CLOUDFLARE_DDNS_HOSTNAME=subdomain.example.com \
  -e CLOUDFLARE_API_TOKEN=your-api-token \
  ghcr.io/jon-frankel/cloudflare-ddns:latest
```

**View logs:**
```bash
docker logs -f cloudflare-ddns
```

## Quick Start

### 1. Initial Setup

Run the binary to start the interactive setup:

```bash
./cloudflare-ddns
```

The setup wizard will ask for:
- Hostname to update (e.g., `home.example.com`)
- Cloudflare API key (input is masked)

The tool validates your credentials before saving.

### 2. Run as a Daemon

Start the polling loop:

```bash
./cloudflare-ddns run
```

This runs a 60-second polling loop that:
- Fetches your current public IP from ipify
- Checks the DNS record in Cloudflare
- Updates only if the IP has changed
- Logs all activity

### 3. Run as a Background Service (macOS)

Using Homebrew's service management:

```bash
brew services start cloudflare-ddns
brew services stop cloudflare-ddns
brew services logs -f cloudflare-ddns
```

## Commands

```bash
cloudflare-ddns                 # First-run setup (if not configured)
cloudflare-ddns run             # Start the polling daemon
cloudflare-ddns test            # Run a single update cycle and show results
cloudflare-ddns logs [-n 50]    # View recent log entries (default last 50 lines)
cloudflare-ddns config          # Manage configuration
cloudflare-ddns --version       # Show version
cloudflare-ddns --help          # Show help
```

### Config Command

Manage your configuration settings:

```bash
# Show current configuration
cloudflare-ddns config show

# Set hostname
cloudflare-ddns config set hostname=new.example.com

# Set API token (prompts for secure input)
cloudflare-ddns config set token

# Set both
cloudflare-ddns config set hostname=new.example.com token
```

### Test Command

Verify your configuration works:

```bash
$ cloudflare-ddns test

Testing configuration for: home.example.com

Hostname:       home.example.com
Current IP:     203.0.113.42
DNS Record IP:  198.51.100.89

✓ DNS record updated successfully!
```

### Logs Command

View recent updates and errors:

```bash
cloudflare-ddns logs -n 100
```

## Configuration

Configuration is stored in:
- **macOS/Linux**: `~/.config/cloudflare-ddns/config.toml`
- **Windows**: `%APPDATA%/cloudflare-ddns/config.toml`

Example config:
```toml
hostname = "home.example.com"
```

API keys are stored securely in:
- **macOS**: Keychain
- **Linux**: Secret Service
- **Windows**: Windows Credential Manager

## Logging

Logs are written to:
- **macOS**: `~/Library/Logs/cloudflare-ddns/`
- **Linux**: `~/.cache/cloudflare-ddns/`

Log files rotate automatically:
- Maximum file size: 10 MB
- Backup files: 3
- Compressed archives enabled

## Behavior

### IP Change Detection

The tool only updates Cloudflare when the public IP changes. This prevents unnecessary API calls and respects rate limits.

### Error Handling

- **Configuration errors**: Logged and exit
- **Transient API failures**: Logged and retried on next cycle
- **Invalid credentials**: Logged as authentication failure, retried next cycle

### First-Run Check

On startup, if either config or keychain entry is missing, interactive setup runs automatically. Once configured, the tool continues to the specified command or defaults to showing configuration status.

## Requirements

- Go 1.24+ (for building from source)
- Cloudflare account with API token
- Internet connection for IP detection and API calls

## Development

### Build

```bash
go build -o cloudflare-ddns
```

### Run Tests

```bash
go test ./...
```

### Run with Verbose Output

```bash
GOFLAGS="-v" go test ./...
```

## Architecture

### Packages

- **cmd/**: Cobra CLI commands (root, run, test, logs)
- **internal/config/**: Configuration file management
- **internal/keychain/**: System keychain integration
- **internal/ip/**: Public IP detection with caching
- **internal/cloudflare/**: Cloudflare API client
- **internal/updater/**: Update orchestration logic
- **internal/logger/**: Structured logging setup

### Update Flow

```
Fetch Public IP (ipify)
    ↓
Compare with Cached IP
    ↓ (if changed)
Get Cloudflare Record
    ↓
Compare with Current IP
    ↓ (if different)
Update Cloudflare
    ↓
Log Result
```

## Security

- API keys stored in system keychain (encrypted, OS-managed)
- Never logs API keys or sensitive credentials
- Password input masked during setup
- Validates credentials before saving configuration

## License

MIT

## Support

For issues, feature requests, or contributions, visit: https://github.com/jon-frankel/cloudflare-ddns
