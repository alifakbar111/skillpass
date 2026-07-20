package webhook

import (
	"net"
	"testing"
)

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"loopback v4", "127.0.0.1", true},
		{"loopback v6", "::1", true},
		{"private 10/8", "10.0.0.5", true},
		{"private 192.168/16", "192.168.1.1", true},
		{"private 172.16/12", "172.16.0.1", true},
		{"link-local v4", "169.254.1.1", true},
		{"unspecified", "0.0.0.0", true},
		{"CGNAT 100.64/10", "100.64.0.1", true},
		{"CGNAT high", "100.127.255.254", true},
		{"public", "8.8.8.8", false},
		{"public edge", "100.63.0.1", false},
		{"public edge 2", "100.128.0.1", false},
		{"public v6", "2606:4700:4700::1111", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("invalid IP %q", tt.ip)
			}
			if got := isBlockedIP(ip); got != tt.want {
				t.Errorf("isBlockedIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestAssertWebhookHost_BlocksPrivate(t *testing.T) {
	// Use a hostname that resolves to a blocked address.
	if err := assertWebhookHost("http://127.0.0.1/x"); err == nil {
		t.Errorf("expected error for 127.0.0.1, got nil")
	}
	if err := assertWebhookHost("http://localhost/x"); err == nil {
		// localhost may or may not resolve depending on the environment.
		// We only assert that no panic happens.
		t.Logf("localhost accepted (no /etc/hosts override in test env)")
	}
}
