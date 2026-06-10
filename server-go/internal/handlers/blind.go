package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// CompanyBlindMode reports whether the given company has blind screening enabled.
// Reads blind_mode via raw SQL (column not part of go-jet generated types).
func CompanyBlindMode(ctx context.Context, db *sql.DB, companyID string) bool {
	var blind bool
	err := db.QueryRowContext(ctx,
		`SELECT blind_mode FROM companies WHERE id = $1`, companyID,
	).Scan(&blind)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// Fail closed-open: treat as non-blind on error (column always exists post-migration).
			return false
		}
		return false
	}
	return blind
}

// MaskCandidateName returns an anonymized display name derived from a profile/candidate id.
func MaskCandidateName(id string) string {
	short := id
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("Candidate %s", short)
}
