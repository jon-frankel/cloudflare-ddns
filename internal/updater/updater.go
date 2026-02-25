package updater

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/jonathan/cloudflare-ddns/internal/cloudflare"
	"github.com/jonathan/cloudflare-ddns/internal/ip"
	"github.com/jonathan/cloudflare-ddns/internal/keychain"
)

type UpdateResult struct {
	CurrentIP     net.IP
	RecordIP      net.IP
	RecordProxied *bool
	Updated       bool
	Error         error
}

// RunOnce performs a single update cycle: fetch public IP, compare with Cloudflare record, update if needed.
func RunOnce(ctx context.Context, hostname string) UpdateResult {
	result := UpdateResult{}

	// Get current public IP
	currentIP, err := ip.Get()
	if err != nil {
		result.Error = fmt.Errorf("failed to get public IP: %w", err)
		slog.Error("Failed to get public IP", "error", err)
		return result
	}
	result.CurrentIP = currentIP

	// Get API token from keychain
	token, err := keychain.Get()
	if err != nil {
		result.Error = fmt.Errorf("failed to get API token from keychain: %w", err)
		slog.Error("Failed to get API token", "error", err)
		return result
	}

	// Create Cloudflare client
	cfClient, err := cloudflare.New(token)
	if err != nil {
		result.Error = fmt.Errorf("failed to create Cloudflare client: %w", err)
		slog.Error("Failed to create Cloudflare client", "error", err)
		return result
	}

	// Get current DNS record
	record, err := cfClient.GetRecord(ctx, hostname)
	if err != nil {
		result.Error = fmt.Errorf("failed to get DNS record: %w", err)
		slog.Error("Failed to get DNS record", "error", err, "hostname", hostname)
		return result
	}
	result.RecordIP = record.IP
	result.RecordProxied = record.Proxied

	// Check if IP or proxy status needs update
	ipNeedsUpdate := !currentIP.Equal(record.IP)
	proxyDisabled := record.Proxied == nil || !*record.Proxied
	needsUpdate := ipNeedsUpdate || proxyDisabled

	if !needsUpdate {
		slog.Info("DNS record is already up to date", "hostname", hostname, "ip", currentIP.String())
		return result
	}

	// Update the record
	updatedRecord, err := cfClient.UpdateRecord(ctx, hostname, currentIP)
	if err != nil {
		result.Error = fmt.Errorf("failed to update DNS record: %w", err)
		slog.Error("Failed to update DNS record", "error", err, "hostname", hostname, "oldIP", record.IP.String(), "newIP", currentIP.String())
		return result
	}

	// Update result with the latest record state
	result.RecordIP = updatedRecord.IP
	result.RecordProxied = updatedRecord.Proxied

	result.Updated = true
	if ipNeedsUpdate {
		slog.Info("Successfully updated DNS record IP", "hostname", hostname, "oldIP", record.IP.String(), "newIP", currentIP.String())
	} else {
		slog.Info("Successfully enabled Cloudflare proxy for DNS record", "hostname", hostname, "ip", currentIP.String())
	}
	return result
}

// RunOnceWithCreate performs a single update cycle, creating the DNS record if it doesn't exist.
// This is useful during initial setup when the record may not have been created yet.
func RunOnceWithCreate(ctx context.Context, hostname string) UpdateResult {
	result := UpdateResult{}

	// Get current public IP
	currentIP, err := ip.Get()
	if err != nil {
		result.Error = fmt.Errorf("failed to get public IP: %w", err)
		slog.Error("Failed to get public IP", "error", err)
		return result
	}
	result.CurrentIP = currentIP

	// Get API token from keychain
	token, err := keychain.Get()
	if err != nil {
		result.Error = fmt.Errorf("failed to get API token from keychain: %w", err)
		slog.Error("Failed to get API token", "error", err)
		return result
	}

	// Create Cloudflare client
	cfClient, err := cloudflare.New(token)
	if err != nil {
		result.Error = fmt.Errorf("failed to create Cloudflare client: %w", err)
		slog.Error("Failed to create Cloudflare client", "error", err)
		return result
	}

	// Get or create DNS record
	record, err := cfClient.GetRecordOrCreate(ctx, hostname, currentIP)
	if err != nil {
		result.Error = fmt.Errorf("failed to get or create DNS record: %w", err)
		slog.Error("Failed to get or create DNS record", "error", err, "hostname", hostname)
		return result
	}
	result.RecordIP = record.IP
	result.RecordProxied = record.Proxied

	// Check if IP or proxy status needs update
	ipNeedsUpdate := !currentIP.Equal(record.IP)
	proxyDisabled := record.Proxied == nil || !*record.Proxied
	needsUpdate := ipNeedsUpdate || proxyDisabled

	if !needsUpdate {
		slog.Info("DNS record is already up to date", "hostname", hostname, "ip", currentIP.String())
		return result
	}

	// Update the record with the current IP
	updatedRecord, err := cfClient.UpdateRecord(ctx, hostname, currentIP)
	if err != nil {
		result.Error = fmt.Errorf("failed to update DNS record: %w", err)
		slog.Error("Failed to update DNS record", "error", err, "hostname", hostname, "oldIP", record.IP.String(), "newIP", currentIP.String())
		return result
	}

	// Update result with the latest record state
	result.RecordIP = updatedRecord.IP
	result.RecordProxied = updatedRecord.Proxied

	result.Updated = true
	if ipNeedsUpdate {
		slog.Info("Successfully updated DNS record IP", "hostname", hostname, "oldIP", record.IP.String(), "newIP", currentIP.String())
	} else {
		slog.Info("Successfully enabled Cloudflare proxy for DNS record", "hostname", hostname, "ip", currentIP.String())
	}
	return result
}
