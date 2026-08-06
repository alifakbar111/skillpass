package identity

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrBadProvider = errors.New("unknown identity provider")
	ErrBadInput    = errors.New("invalid input")
)

type Service struct {
	db     *sql.DB
	signer *Signer
}

func NewService(db *sql.DB, signer *Signer) *Service {
	return &Service{db: db, signer: signer}
}

// DIDResponse is the API/public shape of a DID.
type DIDResponse struct {
	ID        string `json:"id"`
	DID       string `json:"did"`
	PublicKey string `json:"publicKey"`
	Algorithm string `json:"algorithm"`
	IssuerDID string `json:"issuerDid"`
	CreatedAt string `json:"createdAt"`
} //@name DIDResponse

// IssueDID returns the employee's DID, creating one (with a fresh Ed25519
// keypair) on first call. Company-scoped.
func (s *Service) IssueDID(ctx context.Context, companyID, employeeID uuid.UUID) (*DIDResponse, error) {
	var ok bool
	if err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM employees WHERE id=$1 AND company_id=$2)`, employeeID, companyID).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}

	if existing, err := s.GetDID(ctx, companyID, employeeID); err == nil {
		return existing, nil
	}

	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	did := "did:skillpass:" + uuid.New().String()
	pubB64 := base64.StdEncoding.EncodeToString(pub)

	var id uuid.UUID
	var created time.Time
	if err := s.db.QueryRowContext(ctx,
		`INSERT INTO did_identities (employee_id, did_string, public_key, algorithm)
		 VALUES ($1,$2,$3,'ed25519') RETURNING id, created_at`,
		employeeID, did, pubB64).Scan(&id, &created); err != nil {
		return nil, err
	}
	return &DIDResponse{ID: id.String(), DID: did, PublicKey: pubB64, Algorithm: "ed25519",
		IssuerDID: IssuerDID, CreatedAt: created.Format(time.RFC3339)}, nil
}

func (s *Service) GetDID(ctx context.Context, companyID, employeeID uuid.UUID) (*DIDResponse, error) {
	var r DIDResponse
	var created time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT d.id, d.did_string, d.public_key, d.algorithm, d.created_at
		 FROM did_identities d JOIN employees e ON e.id = d.employee_id
		 WHERE d.employee_id=$1 AND e.company_id=$2`, employeeID, companyID).
		Scan(&r.ID, &r.DID, &r.PublicKey, &r.Algorithm, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.IssuerDID = IssuerDID
	r.CreatedAt = created.Format(time.RFC3339)
	return &r, nil
}

// Resolve is the public DID-document lookup by record id.
func (s *Service) Resolve(ctx context.Context, id uuid.UUID) (*DIDResponse, error) {
	var r DIDResponse
	var created time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT id, did_string, public_key, algorithm, created_at FROM did_identities WHERE id=$1 AND revoked_at IS NULL`, id).
		Scan(&r.ID, &r.DID, &r.PublicKey, &r.Algorithm, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.IssuerDID = IssuerDID
	r.CreatedAt = created.Format(time.RFC3339)
	return &r, nil
}

// CredentialResponse is a signed credential with a live verification result.
type CredentialResponse struct {
	ID              string `json:"id"`
	CredentialType  string `json:"credentialType"`
	IssuerDID       string `json:"issuerDid"`
	SubjectDataHash string `json:"subjectDataHash"`
	Signature       string `json:"signature"`
	Algorithm       string `json:"algorithm"`
	IssuedAt        string `json:"issuedAt"`
	Verified        bool   `json:"verified"`
} //@name CredentialResponse

// ListCredentials returns an employee's signed credentials, each re-verified
// against the issuer public key (proving authenticity without a ledger).
func (s *Service) ListCredentials(ctx context.Context, companyID, employeeID uuid.UUID) ([]CredentialResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT sc.id, sc.credential_type, sc.issuer_did, sc.subject_data_hash, sc.signature, sc.algorithm, sc.issued_at
		FROM signed_credentials sc
		JOIN did_identities d ON d.id = sc.did_id
		JOIN employees e ON e.id = d.employee_id
		WHERE d.employee_id=$1 AND e.company_id=$2
		ORDER BY sc.issued_at DESC`, employeeID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issuerPub := s.signer.PublicKeyBase64()
	out := []CredentialResponse{}
	for rows.Next() {
		var cr CredentialResponse
		var issued time.Time
		if err := rows.Scan(&cr.ID, &cr.CredentialType, &cr.IssuerDID, &cr.SubjectDataHash, &cr.Signature, &cr.Algorithm, &issued); err != nil {
			return nil, err
		}
		cr.IssuedAt = issued.Format(time.RFC3339)
		cr.Verified = Verify(issuerPub, cr.Signature, []byte(cr.SubjectDataHash))
		out = append(out, cr)
	}
	return out, rows.Err()
}

// IssueCredential signs a claim about a DID and stores it.
func (s *Service) IssueCredential(ctx context.Context, didID uuid.UUID, credType string, claim []byte, expiresAt *time.Time) (string, error) {
	hash := HashHex(claim)
	sig := s.signer.Sign([]byte(hash))
	var id uuid.UUID
	if err := s.db.QueryRowContext(ctx,
		`INSERT INTO signed_credentials (did_id, credential_type, issuer_did, subject_data_hash, signature, algorithm, expires_at)
		 VALUES ($1,$2,$3,$4,$5,'ed25519',$6) RETURNING id`,
		didID, credType, IssuerDID, hash, sig, expiresAt).Scan(&id); err != nil {
		return "", err
	}
	return id.String(), nil
}

// Anchor appends a signed, hash-chained integrity record for an entity.
func (s *Service) Anchor(ctx context.Context, entityType string, entityID uuid.UUID, sha256hash string) error {
	var prev sql.NullString
	_ = s.db.QueryRowContext(ctx,
		`SELECT sha256_hash FROM integrity_anchors ORDER BY created_at DESC LIMIT 1`).Scan(&prev)

	ts := time.Now().Format(time.RFC3339Nano)
	msg := sha256hash + "|" + ts
	if prev.Valid {
		msg += "|" + prev.String
	}
	sig := s.signer.Sign([]byte(msg))

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO integrity_anchors (entity_type, entity_id, sha256_hash, signature, signed_by, prev_anchor_hash)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		entityType, entityID, sha256hash, sig, IssuerDID, nullStr(prev))
	return err
}

func nullStr(n sql.NullString) any {
	if n.Valid {
		return n.String
	}
	return nil
}

// IssuerPublicKey exposes the issuer's public key (for the future JWKS endpoint).
func (s *Service) IssuerPublicKey() string { return s.signer.PublicKeyBase64() }

// Sanity guard so callers know if the signer is set.
func (s *Service) Ready() bool {
	return s.signer != nil && fmt.Sprint(s.signer.PublicKeyBase64()) != ""
}
