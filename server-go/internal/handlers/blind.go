package handlers

import (
	"context"
	"database/sql"
	"fmt"
)

// CompanyBlindMode reports whether the given company has blind screening enabled.
// Reads blind_mode via raw SQL (column not part of go-jet generated types).
// Any error (including no rows) is treated as blind mode off.
func CompanyBlindMode(ctx context.Context, db *sql.DB, companyID string) bool {
	var blind bool
	if err := db.QueryRowContext(ctx,
		`SELECT blind_mode FROM companies WHERE id = $1`, companyID,
	).Scan(&blind); err != nil {
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
