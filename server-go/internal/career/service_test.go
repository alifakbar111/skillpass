package career

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func TestListPaths(t *testing.T) {
	db := testutil.SetupTestDB()
	svc := NewService(db)

	paths, err := svc.ListPaths(context.Background(), "Technology")
	if err != nil {
		t.Fatalf("list paths: %v", err)
	}
	// Paths may be empty if no seed data, but shouldn't error
	if paths == nil {
		paths = []CareerPath{}
	}
	t.Logf("found %d career paths for Technology", len(paths))
}

func TestGetSkillGap(t *testing.T) {
	db := testutil.SetupTestDB()

	_, pID, _ := testutil.CreateJobseeker(db, "careerjs@ex.com", "careerjs", "pass123", "Career JS")
	testutil.CreateExperience(db, pID, "work", "Junior Developer", "TechCorp")
	svc := NewService(db)

	t.Run("get skill gap", func(t *testing.T) {
		gap, err := svc.GetSkillGap(context.Background(), pID.String(), "Technology")
		if err != nil {
			t.Fatalf("skill gap: %v", err)
		}
		if gap == nil {
			t.Fatal("expected skill gap result")
		}
	})

	t.Run("get skill gap for nonexistent profile", func(t *testing.T) {
		_, err := svc.GetSkillGap(context.Background(), uuid.New().String(), "Technology")
		if err != nil {
			t.Fatalf("skill gap nonexistent: %v", err)
		}
	})
}

func TestPredictPath(t *testing.T) {
	db := testutil.SetupTestDB()

	_, pID, _ := testutil.CreateJobseeker(db, "predjs@ex.com", "predjs", "pass123", "Pred JS")
	testutil.CreateExperience(db, pID, "work", "Developer", "TechCo")
	svc := NewService(db)

	pred, err := svc.PredictPath(context.Background(), pID.String(), "Technology")
	if err != nil {
		t.Fatalf("predict: %v", err)
	}
	if pred == nil {
		t.Fatal("expected prediction")
	}
}
