package profileviews

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func TestRecordView(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "viewco@ex.com", "viewco", "pass123", "View Co", true)
	_, pID, _ := testutil.CreateJobseeker(db, "viewjs@ex.com", "viewjs", "pass123", "View JS")
	viewerID, _ := testutil.CreateUser(db, "viewer@ex.com", "viewer", "pass123", "Viewer", "company")
	svc := NewService(db)

	t.Run("record view", func(t *testing.T) {
		err := svc.RecordView(context.Background(), pID, viewerID, &cID)
		if err != nil {
			t.Fatalf("record: %v", err)
		}
	})

	t.Run("duplicate view deduplicated per day", func(t *testing.T) {
		err := svc.RecordView(context.Background(), pID, viewerID, &cID)
		if err != nil {
			t.Fatalf("record duplicate: %v", err)
		}
		views, _ := svc.GetViewsByProfile(context.Background(), pID)
		if len(views) != 1 {
			t.Fatalf("expected 1 view, got %d", len(views))
		}
	})

	t.Run("different company records separate view", func(t *testing.T) {
		_, cID2, _ := testutil.CreateCompanyUser(db, "viewco2@ex.com", "viewco2", "pass123", "View Co 2", true)
		err := svc.RecordView(context.Background(), pID, viewerID, &cID2)
		if err != nil {
			t.Fatalf("record different company: %v", err)
		}
		views, _ := svc.GetViewsByProfile(context.Background(), pID)
		if len(views) != 2 {
			t.Fatalf("expected 2 views, got %d", len(views))
		}
	})
}

func TestGetViewsByProfile(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "vcoco@ex.com", "vcoco", "pass123", "VCo Co", true)
	_, cID2, _ := testutil.CreateCompanyUser(db, "vcoco2@ex.com", "vcoco2", "pass123", "VCo Co 2", true)
	_, pID, _ := testutil.CreateJobseeker(db, "vcjs@ex.com", "vcjs", "pass123", "VC JS")
	viewer1, _ := testutil.CreateUser(db, "v1@ex.com", "v1", "pass123", "V1", "company")
	viewer2, _ := testutil.CreateUser(db, "v2@ex.com", "v2", "pass123", "V2", "company")
	svc := NewService(db)

	svc.RecordView(context.Background(), pID, viewer1, &cID)
	svc.RecordView(context.Background(), pID, viewer2, &cID2)

	views, err := svc.GetViewsByProfile(context.Background(), pID)
	if err != nil {
		t.Fatalf("views: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
}

func TestGetViewsByProfileEmpty(t *testing.T) {
	db := testutil.SetupTestDB()

	svc := NewService(db)
	views, err := svc.GetViewsByProfile(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("views: %v", err)
	}
	if len(views) != 0 {
		t.Fatalf("expected 0, got %d", len(views))
	}
}
