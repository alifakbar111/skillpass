package identity

import (
	"context"
	"strings"
)

// PDDiktiProvider verifies higher-education records against PDDikti.
//
// STUB: production would query the PDDikti API by name (and ideally NIM/NIK).
// Here a non-empty name passes, so the flow is exercisable without external
// credentials. Swap Verify's body for a real HTTP call.
type PDDiktiProvider struct{}

func (PDDiktiProvider) Name() string { return "pddikti" }

func (PDDiktiProvider) Verify(_ context.Context, in VerifyInput) (status, detail string, err error) {
	if strings.TrimSpace(in.FullName) == "" {
		return "failed", "Name is required for PDDikti verification", nil
	}
	return "verified", "Education record located in PDDikti (stub)", nil
}
