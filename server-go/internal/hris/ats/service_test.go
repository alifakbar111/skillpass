package ats_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"skillpass-server-go/internal/hris/ats"
	"skillpass-server-go/internal/rbac"
	"skillpass-server-go/internal/testutil"
)

func TestPipelineDefaultAndMove(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()
	svc := ats.NewService(db)

	_, companyID, err := testutil.CreateCompanyUser(db, testutil.UniqueUsername("co")+"@ex.com", testutil.UniqueUsername("co"), "pw", "Acme", true)
	if err != nil {
		t.Fatalf("create company: %v", err)
	}

	pipe, err := svc.EnsureDefaultPipeline(ctx, companyID)
	if err != nil {
		t.Fatalf("ensure default pipeline: %v", err)
	}
	if len(pipe.Stages) != 7 {
		t.Fatalf("expected 7 default stages, got %d", len(pipe.Stages))
	}
	// Idempotent: a second call returns the same pipeline, not a new one.
	pipe2, err := svc.EnsureDefaultPipeline(ctx, companyID)
	if err != nil {
		t.Fatalf("ensure default pipeline (2): %v", err)
	}
	if pipe2.ID != pipe.ID {
		t.Fatalf("expected same default pipeline, got %s vs %s", pipe.ID, pipe2.ID)
	}

	cand, err := svc.AddCandidate(ctx, companyID, ats.AddCandidateInput{
		CandidateName:  "Jane Doe",
		CandidateEmail: "jane@ex.com",
	})
	if err != nil {
		t.Fatalf("add candidate: %v", err)
	}
	if cand.CurrentStageID == nil || *cand.CurrentStageID != pipe.Stages[0].ID {
		t.Fatalf("candidate should start at first stage")
	}

	// Move to the third stage.
	targetStage := uuid.MustParse(pipe.Stages[2].ID)
	candID := uuid.MustParse(cand.ID)
	moved, err := svc.MoveCandidate(ctx, companyID, candID, targetStage)
	if err != nil {
		t.Fatalf("move candidate: %v", err)
	}
	if moved.CurrentStageID == nil || *moved.CurrentStageID != pipe.Stages[2].ID {
		t.Fatalf("candidate did not move to target stage")
	}
}

// TestOfferAcceptBridgesToExistingAccount is the Sprint 5 DoD: accepting an
// offer creates an employee record LINKED to the candidate's existing login.
func TestOfferAcceptBridgesToExistingAccount(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()
	svc := ats.NewService(db)
	rbacSvc := rbac.NewService(db)

	_, companyID, err := testutil.CreateCompanyUser(db, testutil.UniqueUsername("co")+"@ex.com", testutil.UniqueUsername("co"), "pw", "Acme", true)
	if err != nil {
		t.Fatalf("create company: %v", err)
	}
	// Seed the company's system roles so "Employee" exists for the bridge.
	if err := rbacSvc.EnsureCompanyRoles(ctx, companyID); err != nil {
		t.Fatalf("ensure roles: %v", err)
	}

	// The candidate already has a marketplace login.
	jsEmail := testutil.UniqueUsername("js") + "@ex.com"
	jsUserID, _, err := testutil.CreateJobseeker(db, jsEmail, testutil.UniqueUsername("js"), "pw", "John Seeker")
	if err != nil {
		t.Fatalf("create jobseeker: %v", err)
	}

	cand, err := svc.AddCandidate(ctx, companyID, ats.AddCandidateInput{
		CandidateName:  "John Seeker",
		CandidateEmail: jsEmail,
	})
	if err != nil {
		t.Fatalf("add candidate: %v", err)
	}
	candID := uuid.MustParse(cand.ID)

	offer, err := svc.CreateOffer(ctx, companyID, candID, jsUserID, ats.OfferInput{
		PositionTitle: "Software Engineer",
		Body:          "We are pleased to offer you the role of Software Engineer.",
	})
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if offer.AcceptToken == "" {
		t.Fatal("offer should carry an accept token")
	}
	if _, err := svc.SendOffer(ctx, companyID, uuid.MustParse(offer.ID)); err != nil {
		t.Fatalf("send offer: %v", err)
	}

	res, err := svc.AcceptOffer(ctx, offer.AcceptToken, "John Seeker")
	if err != nil {
		t.Fatalf("accept offer: %v", err)
	}
	if res.Status != "accepted" {
		t.Fatalf("expected accepted, got %s", res.Status)
	}
	if !res.EmployeeLinked {
		t.Fatal("expected the offer acceptance to link the existing login")
	}

	// Verify: an employee row exists, linked to the jobseeker's user id.
	var linkedUserID uuid.NullUUID
	var status string
	err = db.QueryRowContext(ctx,
		`SELECT user_id, employment_status::text FROM employees WHERE company_id=$1 AND lower(email)=lower($2)`,
		companyID, jsEmail).Scan(&linkedUserID, &status)
	if err != nil {
		t.Fatalf("query employee: %v", err)
	}
	if !linkedUserID.Valid || linkedUserID.UUID != jsUserID {
		t.Fatalf("employee.user_id should equal the jobseeker's user id; got %v want %v", linkedUserID, jsUserID)
	}
	if status != "active" {
		t.Fatalf("new employee should be active, got %s", status)
	}

	// The candidate is now hired.
	cand2, err := svc.GetCandidate(ctx, companyID, candID)
	if err != nil {
		t.Fatalf("get candidate: %v", err)
	}
	if cand2.Status != "hired" {
		t.Fatalf("candidate should be hired, got %s", cand2.Status)
	}

	// Accepting again is idempotent (no error, still accepted).
	if _, err := svc.AcceptOffer(ctx, offer.AcceptToken, "John Seeker"); err != nil {
		t.Fatalf("re-accept should be idempotent: %v", err)
	}
}
