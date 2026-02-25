package cloudflare

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
)

type Client struct {
	api *cf.API
}

type DNSRecord struct {
	ID      string
	Name    string
	IP      net.IP
	TTL     int
	Proxied *bool
}

// New creates a new Cloudflare client with the given API token.
func New(apiToken string) (*Client, error) {
	api, err := cf.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare client: %w", err)
	}
	return &Client{api: api}, nil
}

// GetRecord fetches the A record for the given hostname.
// It automatically extracts the zone (root domain) from the hostname.
func (c *Client) GetRecord(ctx context.Context, hostname string) (*DNSRecord, error) {
	zoneID, err := c.getZoneID(ctx, hostname)
	if err != nil {
		return nil, err
	}

	// Create ResourceContainer for the zone
	rc := cf.ZoneIdentifier(zoneID)

	// List DNS records filtered by name and type A
	records, _, err := c.api.ListDNSRecords(ctx, rc, cf.ListDNSRecordsParams{
		Name: hostname,
		Type: "A",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("A record not found for %s", hostname)
	}

	rec := records[0]
	ip := net.ParseIP(rec.Content)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP in DNS record: %s", rec.Content)
	}

	return &DNSRecord{
		ID:      rec.ID,
		Name:    rec.Name,
		IP:      ip,
		TTL:     rec.TTL,
		Proxied: rec.Proxied,
	}, nil
}

// UpdateRecord updates the A record with a new IP address.
// Always enables Cloudflare proxy (orange cloud) if not already enabled.
// Returns the updated record.
func (c *Client) UpdateRecord(ctx context.Context, hostname string, newIP net.IP) (*DNSRecord, error) {
	zoneID, err := c.getZoneID(ctx, hostname)
	if err != nil {
		return nil, err
	}

	record, err := c.GetRecord(ctx, hostname)
	if err != nil {
		return nil, err
	}

	// Create ResourceContainer for the zone
	rc := cf.ZoneIdentifier(zoneID)

	// Always enable Cloudflare proxy (orange cloud)
	proxied := cf.BoolPtr(true)

	updateParams := cf.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    "A",
		Name:    hostname,
		Content: newIP.String(),
		TTL:     record.TTL,
		Proxied: proxied,
	}

	slog.Debug("Updating DNS record", "hostname", hostname, "oldIP", record.IP.String(), "newIP", newIP.String(), "proxied", true)

	updatedRec, err := c.api.UpdateDNSRecord(ctx, rc, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %w", err)
	}

	slog.Debug("DNS record updated successfully", "hostname", hostname, "newIP", newIP.String())

	// Parse the updated record
	updatedIP := net.ParseIP(updatedRec.Content)
	if updatedIP == nil {
		return nil, fmt.Errorf("invalid IP in updated DNS record: %s", updatedRec.Content)
	}

	return &DNSRecord{
		ID:      updatedRec.ID,
		Name:    updatedRec.Name,
		IP:      updatedIP,
		TTL:     updatedRec.TTL,
		Proxied: updatedRec.Proxied,
	}, nil
}

// CreateRecord creates a new A record with the given IP address.
// Returns the created record.
func (c *Client) CreateRecord(ctx context.Context, hostname string, ip net.IP) (*DNSRecord, error) {
	zoneID, err := c.getZoneID(ctx, hostname)
	if err != nil {
		return nil, err
	}

	// Create ResourceContainer for the zone
	rc := cf.ZoneIdentifier(zoneID)

	createParams := cf.CreateDNSRecordParams{
		Type:    "A",
		Name:    hostname,
		Content: ip.String(),
		TTL:     3600,             // Default TTL of 1 hour
		Proxied: cf.BoolPtr(true), // Enable Cloudflare proxy (orange cloud)
	}

	slog.Debug("Creating DNS record", "hostname", hostname, "ip", ip.String(), "proxied", true)

	rec, err := c.api.CreateDNSRecord(ctx, rc, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	slog.Debug("DNS record created", "id", rec.ID, "proxied", rec.Proxied)

	createdIP := net.ParseIP(rec.Content)
	if createdIP == nil {
		return nil, fmt.Errorf("invalid IP in created DNS record: %s", rec.Content)
	}

	return &DNSRecord{
		ID:      rec.ID,
		Name:    rec.Name,
		IP:      createdIP,
		TTL:     rec.TTL,
		Proxied: rec.Proxied,
	}, nil
}

// GetRecordOrCreate fetches the A record for the given hostname.
// If the record doesn't exist, it creates one with the given IP.
// This is useful during initial setup.
func (c *Client) GetRecordOrCreate(ctx context.Context, hostname string, ip net.IP) (*DNSRecord, error) {
	record, err := c.GetRecord(ctx, hostname)
	if err == nil {
		return record, nil
	}

	// Record doesn't exist, create it
	return c.CreateRecord(ctx, hostname, ip)
}

// getZoneID extracts the root domain from the hostname and fetches its zone ID.
func (c *Client) getZoneID(ctx context.Context, hostname string) (string, error) {
	// Extract the root domain (e.g., "example.com" from "home.example.com")
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid hostname: %s", hostname)
	}

	// Try progressively shorter domain names to find the zone
	for i := 1; i < len(parts); i++ {
		zone := strings.Join(parts[i:], ".")
		zoneID, err := c.api.ZoneIDByName(zone)
		if err == nil {
			return zoneID, nil
		}
	}

	return "", fmt.Errorf("zone not found for %s", hostname)
}
