package ip

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockTransport struct {
	response *http.Response
	err      error
}

func (mt *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if mt.err != nil {
		return nil, mt.err
	}
	return mt.response, nil
}

func TestGetValidIP(t *testing.T) {
	// Reset cache
	cachedIP = nil

	mockClient := &http.Client{
		Transport: &mockTransport{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("192.168.1.1")),
			},
		},
	}

	oldClient := client
	client = mockClient
	defer func() { client = oldClient }()

	ip, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if ip.String() != "192.168.1.1" {
		t.Errorf("Expected IP '192.168.1.1', got '%s'", ip.String())
	}

	// Test caching
	if !IsCached() {
		t.Error("Expected IsCached() to return true after Get()")
	}

	cached := GetCached()
	if !cached.Equal(ip) {
		t.Errorf("Expected cached IP to equal fetched IP")
	}
}

func TestGetInvalidIP(t *testing.T) {
	cachedIP = nil

	mockClient := &http.Client{
		Transport: &mockTransport{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("not-an-ip")),
			},
		},
	}

	oldClient := client
	client = mockClient
	defer func() { client = oldClient }()

	_, err := Get()
	if err == nil {
		t.Error("Expected error for invalid IP")
	}

	if strings.Contains(err.Error(), "invalid IP") {
		t.Logf("Got expected error: %v", err)
	}
}

func TestGetNetworkError(t *testing.T) {
	cachedIP = nil

	mockClient := &http.Client{
		Transport: &mockTransport{
			err: io.EOF,
		},
	}

	oldClient := client
	client = mockClient
	defer func() { client = oldClient }()

	_, err := Get()
	if err == nil {
		t.Error("Expected error for network failure")
	}
}

func TestGetHTTPError(t *testing.T) {
	cachedIP = nil

	mockClient := &http.Client{
		Transport: &mockTransport{
			response: &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader("")),
			},
		},
	}

	oldClient := client
	client = mockClient
	defer func() { client = oldClient }()

	_, err := Get()
	if err == nil {
		t.Error("Expected error for HTTP 500")
	}

	if !strings.Contains(err.Error(), "status") {
		t.Errorf("Expected status error, got: %v", err)
	}
}

func TestGetCachedBeforeFetch(t *testing.T) {
	cachedIP = nil

	if IsCached() {
		t.Error("Expected IsCached() to return false before any fetch")
	}

	if GetCached() != nil {
		t.Error("Expected GetCached() to return nil before any fetch")
	}
}

func TestIPWithWhitespace(t *testing.T) {
	cachedIP = nil

	mockClient := &http.Client{
		Transport: &mockTransport{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("  10.20.30.40  \n")),
			},
		},
	}

	oldClient := client
	client = mockClient
	defer func() { client = oldClient }()

	ip, err := Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if ip.String() != "10.20.30.40" {
		t.Errorf("Expected IP '10.20.30.40', got '%s'", ip.String())
	}
}
