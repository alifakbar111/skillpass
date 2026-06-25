package companyreviews

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"skillpass-server-go/internal/testutil"
)

func TestCreateReview(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "revco@ex.com", "revco", "pass123", "Rev Co", true)
	_, pID, _ := testutil.CreateJobseeker(db, "revjs@ex.com", "revjs", "pass123", "Rev JS")
	svc := NewService(db)

	t.Run("submit review", func(t *testing.T) {
		review, err := svc.Create(context.Background(), cID.String(), pID.String(), CreateReviewRequest{
			Rating:          4,
			Review:          "Good work culture",
			InteractionType: "applied",
		})
		if err != nil {
			t.Fatalf("submit: %v", err)
		}
		if review.Rating != 4 {
			t.Fatalf("expected 4, got %d", review.Rating)
		}
	})

	t.Run("invalid rating rejected", func(t *testing.T) {
		_, err := svc.Create(context.Background(), cID.String(), pID.String(), CreateReviewRequest{
			Rating:          0,
			Review:          "Bad",
			InteractionType: "applied",
		})
		if err == nil {
			t.Fatal("expected error for invalid rating")
		}
	})

	t.Run("invalid interaction type rejected", func(t *testing.T) {
		_, err := svc.Create(context.Background(), cID.String(), pID.String(), CreateReviewRequest{
			Rating:          3,
			Review:          "OK",
			InteractionType: "unknown",
		})
		if err == nil {
			t.Fatal("expected error for invalid interaction type")
		}
	})
}

func TestGetCompanyReviews(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "grco@ex.com", "grco", "pass123", "GR Co", true)
	svc := NewService(db)

	for i := 0; i < 3; i++ {
		_, pID, _ := testutil.CreateJobseeker(db, "grjs"+string(rune('0'+i))+"@ex.com", "grjs"+string(rune('0'+i)), "pass123", "GR JS")
		svc.Create(context.Background(), cID.String(), pID.String(), CreateReviewRequest{
			Rating:          4,
			Review:          "Good",
			InteractionType: "applied",
		})
	}

	reviews, err := svc.ListByCompanyID(context.Background(), cID.String())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(reviews) != 3 {
		t.Fatalf("expected 3 reviews, got %d", len(reviews))
	}
}

func TestGetReputation(t *testing.T) {
	db := testutil.SetupTestDB()

	_, cID, _ := testutil.CreateCompanyUser(db, "repco@ex.com", "repco", "pass123", "Rep Co", true)
	svc := NewService(db)

		svc.Create(context.Background(), cID.String(), uuid.New().String(), CreateReviewRequest{
		Rating:          5,
		Review:          "Excellent",
		InteractionType: "interviewed",
	})

	rep, err := svc.GetReputation(context.Background(), cID.String())
	if err != nil {
		t.Fatalf("reputation: %v", err)
	}
	if rep.ReviewCount != 1 {
		t.Fatalf("expected 1 review, got %d", rep.ReviewCount)
	}
	if rep.AverageRate != 5.0 {
		t.Fatalf("expected 5.0, got %.1f", rep.AverageRate)
	}
}
