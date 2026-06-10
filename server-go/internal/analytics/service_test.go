package analytics

import (
	"context"
	"testing"

	"skillpass-server-go/internal/testutil"
)

func TestCompanyAnalytics(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()

	_, cID, _ := testutil.CreateCompanyUser(db, "ana@ex.com", "anaco", "pass123", "Analytics Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Data Engineer", "Technology", true)
	testutil.CreateJob(db, cID, "Closed Role", "Technology", false)

	_, p1, _ := testutil.CreateJobseeker(db, "ana1@ex.com", "ana1", "pass123", "Cand One")
	_, p2, _ := testutil.CreateJobseeker(db, "ana2@ex.com", "ana2", "pass123", "Cand Two")
	testutil.CreateApplication(db, p1, jID, "applied")
	testutil.CreateApplication(db, p2, jID, "offered")

	svc := NewService(db)
	result, err := svc.ForCompany(ctx, cID.String())
	if err != nil {
		t.Fatalf("ForCompany: %v", err)
	}

	if result.TotalJobs != 2 {
		t.Fatalf("expected 2 jobs, got %d", result.TotalJobs)
	}
	if result.OpenJobs != 1 {
		t.Fatalf("expected 1 open job, got %d", result.OpenJobs)
	}
	if result.TotalApplications != 2 {
		t.Fatalf("expected 2 applications, got %d", result.TotalApplications)
	}
	if result.AvgDaysToDecision == nil {
		t.Fatal("expected avgDaysToDecision for offered application")
	}
	if len(result.Jobs) != 2 {
		t.Fatalf("expected 2 job funnels, got %d", len(result.Jobs))
	}
}

func TestJobseekerAnalytics(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()

	_, cID, _ := testutil.CreateCompanyUser(db, "anb@ex.com", "anbco", "pass123", "Analytics Co B", true)
	j1, _ := testutil.CreateJob(db, cID, "Role A", "Technology", true)
	j2, _ := testutil.CreateJob(db, cID, "Role B", "Technology", true)

	_, pID, _ := testutil.CreateJobseeker(db, "anb1@ex.com", "anb1", "pass123", "Cand B")
	testutil.CreateApplication(db, pID, j1, "applied")
	testutil.CreateApplication(db, pID, j2, "reviewed")

	svc := NewService(db)
	result, err := svc.ForJobseeker(ctx, pID.String())
	if err != nil {
		t.Fatalf("ForJobseeker: %v", err)
	}

	if result.TotalApplications != 2 {
		t.Fatalf("expected 2 applications, got %d", result.TotalApplications)
	}
	if result.ResponseRate == nil || *result.ResponseRate != 50 {
		t.Fatalf("expected 50%% response rate, got %v", result.ResponseRate)
	}
	if result.PassportViews != 0 {
		t.Fatalf("expected 0 passport views, got %d", result.PassportViews)
	}
}

func TestJobseekerAnalyticsEmpty(t *testing.T) {
	db := testutil.SetupTestDB()
	_, pID, _ := testutil.CreateJobseeker(db, "anc@ex.com", "anc", "pass123", "Cand C")

	svc := NewService(db)
	result, err := svc.ForJobseeker(context.Background(), pID.String())
	if err != nil {
		t.Fatalf("ForJobseeker: %v", err)
	}
	if result.TotalApplications != 0 {
		t.Fatalf("expected 0 applications, got %d", result.TotalApplications)
	}
	if result.ResponseRate != nil {
		t.Fatal("expected nil response rate with no applications")
	}
}
