package identity

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Provider is an external identity source.
type Provider interface {
	Name() string
	// Verify checks the employee's identity fields and returns a status
	// ("verified" | "failed") plus a human-readable detail.
	Verify(ctx context.Context, in VerifyInput) (status, detail string, err error)
}

// VerifyInput carries the employee data a provider needs.
type VerifyInput struct {
	NationalID string // NIK (KTP) for Dukcapil
	FullName   string
	Education  string // for PDDikti
}

// VerificationResponse is the API shape of a verification record.
type VerificationResponse struct {
	ID         string  `json:"id"`
	Provider   string  `json:"provider"`
	Status     string  `json:"status"`
	Detail     *string `json:"detail,omitempty"`
	VerifiedAt *string `json:"verifiedAt,omitempty"`
	CreatedAt  string  `json:"createdAt"`
} //@name IdentityVerification

// providers is the registry of available identity providers, keyed by name.
func (s *Service) providers() map[string]Provider {
	return map[string]Provider{
		"dukcapil": DukcapilProvider{},
		"pddikti":  PDDiktiProvider{},
	}
}

// RunVerification triggers an identity check for an employee via a provider and
// records the outcome. Company-scoped. "manual" records a verified pass.
func (s *Service) RunVerification(ctx context.Context, companyID, employeeID uuid.UUID, providerName string) (*VerificationResponse, error) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))

	// Load the employee's identity fields (scoped to company).
	var nationalID, firstName, lastName sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT national_id, first_name, last_name FROM employees WHERE id=$1 AND company_id=$2`,
		employeeID, companyID).Scan(&nationalID, &firstName, &lastName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	status, detail := "verified", "Manually marked verified"
	if providerName != "manual" {
		provider, ok := s.providers()[providerName]
		if !ok {
			return nil, ErrBadProvider
		}
		st, dt, verr := provider.Verify(ctx, VerifyInput{
			NationalID: nationalID.String,
			FullName:   strings.TrimSpace(firstName.String + " " + lastName.String),
		})
		if verr != nil {
			status, detail = "failed", verr.Error()
		} else {
			status, detail = st, dt
		}
	}

	var id uuid.UUID
	var created time.Time
	var verifiedAt *time.Time
	if status == "verified" {
		now := time.Now()
		verifiedAt = &now
	}
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO identity_verifications (employee_id, provider, response_status, detail, verified_at)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		employeeID, providerName, status, detail, verifiedAt).Scan(&id, &created)
	if err != nil {
		return nil, err
	}

	resp := &VerificationResponse{
		ID:        id.String(),
		Provider:  providerName,
		Status:    status,
		Detail:    &detail,
		CreatedAt: created.Format(time.RFC3339),
	}
	if verifiedAt != nil {
		v := verifiedAt.Format(time.RFC3339)
		resp.VerifiedAt = &v
	}
	return resp, nil
}

// ListVerifications returns an employee's verification history (company-scoped).
func (s *Service) ListVerifications(ctx context.Context, companyID, employeeID uuid.UUID) ([]VerificationResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT v.id, v.provider::text, v.response_status::text, v.detail, v.verified_at, v.created_at
		FROM identity_verifications v
		JOIN employees e ON e.id = v.employee_id
		WHERE v.employee_id=$1 AND e.company_id=$2
		ORDER BY v.created_at DESC`, employeeID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []VerificationResponse{}
	for rows.Next() {
		var r VerificationResponse
		var detail *string
		var verifiedAt *time.Time
		var created time.Time
		if err := rows.Scan(&r.ID, &r.Provider, &r.Status, &detail, &verifiedAt, &created); err != nil {
			return nil, err
		}
		r.Detail = detail
		if verifiedAt != nil {
			v := verifiedAt.Format(time.RFC3339)
			r.VerifiedAt = &v
		}
		r.CreatedAt = created.Format(time.RFC3339)
		out = append(out, r)
	}
	return out, rows.Err()
}
