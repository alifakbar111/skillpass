package identity_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"skillpass-server-go/internal/identity"
	"skillpass-server-go/internal/testutil"
)

const testSecret = "identity-test-secret-abcdefghijklmnop"

func TestAttestAndVerifyCredential(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()
	svc := identity.NewService(db, identity.NewSigner(testSecret))

	_, companyID, err := testutil.CreateCompanyUser(db, testutil.UniqueUsername("co")+"@ex.com", testutil.UniqueUsername("co"), "pw", "Acme", true)
	if err != nil {
		t.Fatalf("company: %v", err)
	}
	jsUserID, profileID, err := testutil.CreateJobseeker(db, testutil.UniqueUsername("js")+"@ex.com", testutil.UniqueUsername("js"), "pw", "Jane Dev")
	if err != nil {
		t.Fatalf("jobseeker: %v", err)
	}
	if err := testutil.CreateAIEvaluation(db, profileID, 90); err != nil {
		t.Fatalf("evaluation: %v", err)
	}

	employeeID := uuid.New()
	if _, err := db.ExecContext(ctx, `
		INSERT INTO employees (id, company_id, user_id, employee_id_number, first_name, last_name, email, employment_type, employment_status, join_date)
		VALUES ($1,$2,$3,$4,'Jane','Dev',$5,'permanent','active',$6)`,
		employeeID, companyID, jsUserID, "EMP-"+employeeID.String()[:8], "jane@ex.com", time.Now()); err != nil {
		t.Fatalf("insert employee: %v", err)
	}

	count, err := svc.AttestEmployeeSkills(ctx, companyID, employeeID, nil)
	if err != nil {
		t.Fatalf("attest: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 attested skill, got %d", count)
	}

	atts, err := svc.ListAttestations(ctx, companyID, employeeID)
	if err != nil {
		t.Fatalf("list attestations: %v", err)
	}
	if len(atts) != 1 || !atts[0].Verified {
		t.Fatalf("expected 1 verified attestation, got %+v", atts)
	}

	attID := uuid.MustParse(atts[0].ID)
	vc, err := svc.VerifyCredential(ctx, attID)
	if err != nil {
		t.Fatalf("verify credential: %v", err)
	}
	if !vc.Verified || vc.SkillName != "Go" || vc.Score != 90 {
		t.Fatalf("credential should verify with Go=90, got %+v", vc)
	}

	// Tamper: corrupt the stored signature — verification must now fail.
	if _, err := db.ExecContext(ctx, `UPDATE skill_attestations SET signature='dGFtcGVyZWQ=' WHERE id=$1`, attID); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	vc2, err := svc.VerifyCredential(ctx, attID)
	if err != nil {
		t.Fatalf("verify after tamper: %v", err)
	}
	if vc2.Verified {
		t.Fatal("tampered credential must not verify")
	}

	// JWKS exposes exactly one issuer key.
	jwks := svc.JWKS()
	if len(jwks.Keys) != 1 || jwks.Keys[0].X == "" || jwks.Keys[0].Crv != "Ed25519" {
		t.Fatalf("unexpected JWKS: %+v", jwks)
	}
}

func TestAnchorHashChain(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()
	svc := identity.NewService(db, identity.NewSigner(testSecret))

	id1, id2 := uuid.New(), uuid.New()
	if err := svc.Anchor(ctx, "document", id1, "hashA"); err != nil {
		t.Fatalf("anchor 1: %v", err)
	}
	if err := svc.Anchor(ctx, "document", id2, "hashB"); err != nil {
		t.Fatalf("anchor 2: %v", err)
	}

	// The second anchor links back to the first via prev_anchor_hash.
	var prev string
	if err := db.QueryRowContext(ctx,
		`SELECT prev_anchor_hash FROM integrity_anchors WHERE entity_id=$1`, id2).Scan(&prev); err != nil {
		t.Fatalf("read chain: %v", err)
	}
	if prev != "hashA" {
		t.Fatalf("hash chain broken: prev_anchor_hash=%q want hashA", prev)
	}

	// Tampering the first anchor's hash breaks the link the second one recorded.
	if _, err := db.ExecContext(ctx, `UPDATE integrity_anchors SET sha256_hash='tampered' WHERE entity_id=$1`, id1); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	var cur string
	if err := db.QueryRowContext(ctx, `SELECT sha256_hash FROM integrity_anchors WHERE entity_id=$1`, id1).Scan(&cur); err != nil {
		t.Fatalf("read tampered: %v", err)
	}
	if cur == prev {
		t.Fatal("expected tamper to break the chain link, but hashes still match")
	}
}
