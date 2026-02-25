package ip

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	cachedIP net.IP
	client   = &http.Client{
		Timeout: 10 * time.Second,
	}
)

// Get fetches the current public IP from ipify and caches it.
func Get() (net.IP, error) {
	resp, err := client.Get("https://api.ipify.org?format=text")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IP from ipify: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipify returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read ipify response: %w", err)
	}

	ipStr := strings.TrimSpace(string(body))
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP returned from ipify: %s", ipStr)
	}

	cachedIP = ip
	return ip, nil
}

// GetCached returns the cached IP without making a network call.
func GetCached() net.IP {
	return cachedIP
}

// IsCached returns true if we have a cached IP.
func IsCached() bool {
	return cachedIP != nil
}

// SetClient allows injection of a custom HTTP client (useful for testing).
func SetClient(c *http.Client) {
	client = c
}
