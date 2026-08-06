package identity

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PassportSettingsResponse is the company-side view of an employee's public
// Skill Passport configuration.
type PassportSettingsResponse struct {
	EmployeeID string `json:"employeeId"`
	Slug       string `json:"slug"`
	IsPublic   bool   `json:"isPublic"`
	PublicPath string `json:"publicPath"` // /verify/passport/<slug>
	UpdatedAt  string `json:"updatedAt"`
} //@name PassportSettings

// GetOrCreatePassport returns the employee's passport settings, minting a slug
// on first use. Company-scoped.
func (s *Service) GetOrCreatePassport(ctx context.Context, companyID, employeeID uuid.UUID) (*PassportSettingsResponse, error) {
	var firstName, lastName string
	err := s.db.QueryRowContext(ctx,
		`SELECT first_name, last_name FROM employees WHERE id=$1 AND company_id=$2`, employeeID, companyID).
		Scan(&firstName, &lastName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var slug string
	var isPublic bool
	var updated time.Time
	err = s.db.QueryRowContext(ctx,
		`SELECT public_url_slug, is_public, updated_at FROM skill_passport_public WHERE employee_id=$1`, employeeID).
		Scan(&slug, &isPublic, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		slug = s.uniqueSlug(ctx, firstName, lastName)
		if err = s.db.QueryRowContext(ctx,
			`INSERT INTO skill_passport_public (employee_id, public_url_slug, is_public) VALUES ($1,$2,false)
			 RETURNING public_url_slug, is_public, updated_at`, employeeID, slug).
			Scan(&slug, &isPublic, &updated); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &PassportSettingsResponse{
		EmployeeID: employeeID.String(),
		Slug:       slug,
		IsPublic:   isPublic,
		PublicPath: "/verify/passport/" + slug,
		UpdatedAt:  updated.Format(time.RFC3339),
	}, nil
}

// SetPassportPublic toggles the public visibility of an employee's passport.
func (s *Service) SetPassportPublic(ctx context.Context, companyID, employeeID uuid.UUID, isPublic bool) (*PassportSettingsResponse, error) {
	if _, err := s.GetOrCreatePassport(ctx, companyID, employeeID); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE skill_passport_public sp SET is_public=$3, updated_at=now()
		 FROM employees e WHERE sp.employee_id=e.id AND sp.employee_id=$1 AND e.company_id=$2`,
		employeeID, companyID, isPublic); err != nil {
		return nil, err
	}
	return s.GetOrCreatePassport(ctx, companyID, employeeID)
}

func (s *Service) uniqueSlug(ctx context.Context, firstName, lastName string) string {
	base := slugify(firstName + "-" + lastName)
	if base == "" {
		base = "member"
	}
	for i := 0; i < 5; i++ {
		candidate := base + "-" + randHex(3)
		var exists bool
		if err := s.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM skill_passport_public WHERE public_url_slug=$1)`, candidate).Scan(&exists); err != nil {
			return candidate
		}
		if !exists {
			return candidate
		}
	}
	return base + "-" + randHex(6)
}

// ---------- Public passport ----------

// PublicSkillBadge is one verified skill on a public passport.
type PublicSkillBadge struct {
	AttestationID string `json:"attestationId"`
	SkillName     string `json:"skillName"`
	Score         int    `json:"score"`
	Verified      bool   `json:"verified"`
	VerifyPath    string `json:"verifyPath"` // /verify/credential?id=<attestationId>
} //@name PublicSkillBadge

// PublicPassportResponse is the chain-free public Skill Passport.
type PublicPassportResponse struct {
	Name             string             `json:"name"`
	CompanyName      string             `json:"companyName,omitempty"`
	IssuerDID        string             `json:"issuerDid"`
	IdentityVerified bool               `json:"identityVerified"`
	Skills           []PublicSkillBadge `json:"skills,omitempty"`
} //@name PublicPassport

// GetPublicPassport assembles a public, signature-verified Skill Passport by
// slug. Only returned when the employee has opted in (is_public=true).
func (s *Service) GetPublicPassport(ctx context.Context, slug string) (*PublicPassportResponse, error) {
	var employeeID uuid.UUID
	var firstName, lastName, companyName string
	err := s.db.QueryRowContext(ctx, `
		SELECT sp.employee_id, e.first_name, e.last_name, co.company_name
		FROM skill_passport_public sp
		JOIN employees e ON e.id = sp.employee_id
		JOIN companies co ON co.id = e.company_id
		WHERE sp.public_url_slug=$1 AND sp.is_public=true`, slug).
		Scan(&employeeID, &firstName, &lastName, &companyName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	resp := &PublicPassportResponse{
		Name:        strings.TrimSpace(firstName + " " + lastName),
		CompanyName: companyName,
		IssuerDID:   IssuerDID,
		Skills:      []PublicSkillBadge{},
	}

	// Identity verified if the employee has any successful verification.
	_ = s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM identity_verifications WHERE employee_id=$1 AND response_status='verified')`,
		employeeID).Scan(&resp.IdentityVerified)

	// Active (non-revoked) signed skill attestations.
	issuerPub := s.signer.PublicKeyBase64()
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, skill_name, score, attestation_hash, signature
		FROM skill_attestations
		WHERE employee_id=$1 AND revoked_at IS NULL
		ORDER BY score DESC, skill_name`, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var b PublicSkillBadge
		var hash, sig string
		if err := rows.Scan(&b.AttestationID, &b.SkillName, &b.Score, &hash, &sig); err != nil {
			return nil, err
		}
		b.Verified = Verify(issuerPub, sig, []byte(hash))
		b.VerifyPath = "/verify/credential?id=" + b.AttestationID
		resp.Skills = append(resp.Skills, b)
	}
	return resp, rows.Err()
}

// ---------- helpers ----------

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r == ' ' || r == '-' || r == '_':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func randHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "x"
	}
	return hex.EncodeToString(buf)
}
