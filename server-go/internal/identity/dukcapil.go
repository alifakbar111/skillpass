package identity

import (
	"context"
	"strings"
)

// DukcapilProvider verifies an Indonesian KTP/NIK against Dukcapil.
//
// This is a STUB: production would call the Dukcapil API (or an accredited
// gateway) with the NIK + name. Here we do format validation only — a 16-digit
// NIK passes, anything else fails — so the flow is exercisable end-to-end
// without external credentials. Swap Verify's body for a real HTTP call.
type DukcapilProvider struct{}

func (DukcapilProvider) Name() string { return "dukcapil" }

func (DukcapilProvider) Verify(_ context.Context, in VerifyInput) (status, detail string, err error) {
	nik := strings.TrimSpace(in.NationalID)
	if len(nik) != 16 || !isAllDigits(nik) {
		return "failed", "NIK must be a 16-digit number (stub validation)", nil
	}
	if strings.TrimSpace(in.FullName) == "" {
		return "failed", "Name is required for Dukcapil verification", nil
	}
	return "verified", "NIK format valid; matched against Dukcapil (stub)", nil
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
