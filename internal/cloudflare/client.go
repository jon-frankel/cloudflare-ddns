package cloudflare

import (
	"context"
	"fmt"
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
func (c *Client) UpdateRecord(ctx context.Context, hostname string, newIP net.IP) error {
	zoneID, err := c.getZoneID(ctx, hostname)
	if err != nil {
		return err
	}

	record, err := c.GetRecord(ctx, hostname)
	if err != nil {
		return err
	}

	// Create ResourceContainer for the zone
	rc := cf.ZoneIdentifier(zoneID)

	updateParams := cf.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    "A",
		Name:    hostname,
		Content: newIP.String(),
		TTL:     record.TTL,
		Proxied: record.Proxied,
	}

	_, err = c.api.UpdateDNSRecord(ctx, rc, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
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
