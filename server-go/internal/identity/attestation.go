package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// skillScore mirrors evaluation.SkillScoreItem (read via raw SQL to avoid an
// import cycle with the evaluation package).
type skillScore struct {
	Skill    string `json:"skill"`
	Category string `json:"category"`
	Score    int    `json:"score"`
}

// AttestationClaim is the canonical, signed representation of one skill score.
// The signature covers the SHA-256 of this JSON, so anyone can recompute the
// hash and verify authenticity against the issuer public key (JWKS).
type AttestationClaim struct {
	Type       string `json:"type"`
	IssuerDID  string `json:"issuerDid"`
	EmployeeID string `json:"employeeId"`
	SkillName  string `json:"skillName"`
	Score      int    `json:"score"`
	IssuedAt   string `json:"issuedAt"`
} //@name AttestationClaim

// AttestationResponse is a stored attestation with a live verification result.
type AttestationResponse struct {
	ID        string `json:"id"`
	SkillName string `json:"skillName"`
	Score     int    `json:"score"`
	Hash      string `json:"hash"`
	Signature string `json:"signature"`
	Algorithm string `json:"algorithm"`
	IssuedAt  string `json:"issuedAt"`
	Revoked   bool   `json:"revoked"`
	Verified  bool   `json:"verified"`
} //@name SkillAttestation

// AttestEmployeeSkills signs each skill score from an employee's evaluation and
// stores the attestations. If evaluationID is nil the employee's current
// evaluation is used. Company-scoped. Returns the number of skills attested.
func (s *Service) AttestEmployeeSkills(ctx context.Context, companyID, employeeID uuid.UUID, evaluationID *uuid.UUID) (int, error) {
	// Resolve the employee's evaluation skill_scores JSON.
	var scoresJSON string
	var query string
	args := []any{employeeID, companyID}
	if evaluationID != nil {
		query = `
			SELECT ev.skill_scores
			FROM ai_evaluations ev
			JOIN jobseeker_profiles jp ON jp.id = ev.profile_id
			JOIN employees emp ON emp.user_id = jp.user_id
			WHERE emp.id=$1 AND emp.company_id=$2 AND ev.id=$3`
		args = append(args, *evaluationID)
	} else {
		query = `
			SELECT ev.skill_scores
			FROM ai_evaluations ev
			JOIN jobseeker_profiles jp ON jp.id = ev.profile_id
			JOIN employees emp ON emp.user_id = jp.user_id
			WHERE emp.id=$1 AND emp.company_id=$2 AND ev.is_current = true
			ORDER BY ev.created_at DESC LIMIT 1`
	}
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&scoresJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	var scores []skillScore
	if err := json.Unmarshal([]byte(scoresJSON), &scores); err != nil {
		return 0, err
	}
	if len(scores) == 0 {
		return 0, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Revoke any prior un-revoked attestations for this employee so the passport
	// reflects the latest evaluation only.
	if _, err := tx.ExecContext(ctx,
		`UPDATE skill_attestations SET revoked_at=now() WHERE employee_id=$1 AND revoked_at IS NULL`, employeeID); err != nil {
		return 0, err
	}

	issuedAt := time.Now().UTC().Format(time.RFC3339)
	count := 0
	for _, sc := range scores {
		claim := AttestationClaim{
			Type:       "SkillAttestation",
			IssuerDID:  IssuerDID,
			EmployeeID: employeeID.String(),
			SkillName:  sc.Skill,
			Score:      sc.Score,
			IssuedAt:   issuedAt,
		}
		claimBytes, _ := json.Marshal(claim)
		hash := HashHex(claimBytes)
		sig := s.signer.Sign([]byte(hash))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO skill_attestations (employee_id, skill_name, score, evaluation_id, attestation_hash, signature, algorithm, issued_at)
			VALUES ($1,$2,$3,$4,$5,$6,'ed25519',$7)`,
			employeeID, sc.Skill, sc.Score, evaluationID, hash, sig, issuedAt); err != nil {
			return 0, err
		}
		count++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

// ListAttestations returns an employee's active attestations, each re-verified.
func (s *Service) ListAttestations(ctx context.Context, companyID, employeeID uuid.UUID) ([]AttestationResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT sa.id, sa.skill_name, sa.score, sa.attestation_hash, sa.signature, sa.algorithm, sa.issued_at, sa.revoked_at
		FROM skill_attestations sa
		JOIN employees e ON e.id = sa.employee_id
		WHERE sa.employee_id=$1 AND e.company_id=$2
		ORDER BY sa.issued_at DESC, sa.skill_name`, employeeID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return s.scanAttestations(rows)
}

func (s *Service) scanAttestations(rows *sql.Rows) ([]AttestationResponse, error) {
	issuerPub := s.signer.PublicKeyBase64()
	out := []AttestationResponse{}
	for rows.Next() {
		var a AttestationResponse
		var issued time.Time
		var revoked *time.Time
		if err := rows.Scan(&a.ID, &a.SkillName, &a.Score, &a.Hash, &a.Signature, &a.Algorithm, &issued, &revoked); err != nil {
			return nil, err
		}
		a.IssuedAt = issued.Format(time.RFC3339)
		a.Revoked = revoked != nil
		a.Verified = Verify(issuerPub, a.Signature, []byte(a.Hash))
		out = append(out, a)
	}
	return out, rows.Err()
}

// VerifiedCredential is the public /verify/credential result.
type VerifiedCredential struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	SkillName string `json:"skillName"`
	Score     int    `json:"score"`
	IssuerDID string `json:"issuerDid"`
	IssuedAt  string `json:"issuedAt"`
	Revoked   bool   `json:"revoked"`
	Verified  bool   `json:"verified"`
} //@name VerifiedCredential

// VerifyCredential re-checks a stored attestation's signature against the
// issuer key (public — no company scope). This is the chain-free equivalent of
// "view on-chain": authenticity comes from the Ed25519 signature.
func (s *Service) VerifyCredential(ctx context.Context, attestationID uuid.UUID) (*VerifiedCredential, error) {
	var vc VerifiedCredential
	var hash, sig string
	var issued time.Time
	var revoked *time.Time
	err := s.db.QueryRowContext(ctx, `
		SELECT id, skill_name, score, attestation_hash, signature, issued_at, revoked_at
		FROM skill_attestations WHERE id=$1`, attestationID).
		Scan(&vc.ID, &vc.SkillName, &vc.Score, &hash, &sig, &issued, &revoked)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	vc.Type = "SkillAttestation"
	vc.IssuerDID = IssuerDID
	vc.IssuedAt = issued.Format(time.RFC3339)
	vc.Revoked = revoked != nil
	vc.Verified = !vc.Revoked && Verify(s.signer.PublicKeyBase64(), sig, []byte(hash))
	return &vc, nil
}
