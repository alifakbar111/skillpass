package companyreviews

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCompanyNotFound  = errors.New("company not found")
	ErrProfileNotFound  = errors.New("jobseeker profile not found")
	ErrInvalidRating    = errors.New("rating must be between 1 and 5")
	ErrInvalidInteraction = errors.New("interaction type must be 'applied' or 'interviewed'")
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type CompanyReview struct {
	ID              string `json:"id"`
	CompanyID       string `json:"companyId"`
	CandidateID     string `json:"candidateId"`
	Rating          int    `json:"rating"`
	Review          string `json:"review"`
	InteractionType string `json:"interactionType"`
	CreatedAt       string `json:"createdAt"`
}

type Reputation struct {
	CompanyID   string  `json:"companyId"`
	AverageRate float64 `json:"averageRate"`
	ReviewCount int     `json:"reviewCount"`
}

type CreateReviewRequest struct {
	Rating          int    `json:"rating" binding:"required,min=1,max=5"`
	Review          string `json:"review" binding:"max=5000"`
	InteractionType string `json:"interactionType" binding:"required"`
}

func (s *Service) Create(ctx context.Context, companyID, candidateID string, req CreateReviewRequest) (*CompanyReview, error) {
	if req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRating
	}
	if req.InteractionType != "applied" && req.InteractionType != "interviewed" {
		return nil, ErrInvalidInteraction
	}

	// Check company exists
	var companyExists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1)`, companyID).Scan(&companyExists)
	if err != nil {
		return nil, fmt.Errorf("check company existence: %w", err)
	}
	if !companyExists {
		return nil, ErrCompanyNotFound
	}

	var id string
	var createdAt time.Time
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO company_reviews (company_id, candidate_id, rating, review, interaction_type)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (company_id, candidate_id)
		 DO UPDATE SET rating = EXCLUDED.rating,
		               review = EXCLUDED.review,
		               interaction_type = EXCLUDED.interaction_type,
		               created_at = now()
		 RETURNING id, created_at`,
		companyID, candidateID, req.Rating, req.Review, req.InteractionType,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("upsert review: %w", err)
	}

	return &CompanyReview{
		ID:              id,
		CompanyID:       companyID,
		CandidateID:     candidateID,
		Rating:          req.Rating,
		Review:          req.Review,
		InteractionType: req.InteractionType,
		CreatedAt:       createdAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) GetByID(ctx context.Context, reviewID string) (*CompanyReview, error) {
	var r CompanyReview
	var createdAt time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_id, candidate_id, rating, COALESCE(review, ''), interaction_type, created_at
		 FROM company_reviews
		 WHERE id = $1`,
		reviewID,
	).Scan(&r.ID, &r.CompanyID, &r.CandidateID, &r.Rating, &r.Review, &r.InteractionType, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get review: %w", err)
	}
	r.CreatedAt = createdAt.Format(time.RFC3339)
	return &r, nil
}

func (s *Service) GetReputation(ctx context.Context, companyID string) (*Reputation, error) {
	var avgRate sql.NullFloat64
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(AVG(rating), 0), COUNT(*)
		 FROM company_reviews
		 WHERE company_id = $1`,
		companyID,
	).Scan(&avgRate, &count)
	if err != nil {
		return nil, fmt.Errorf("get reputation: %w", err)
	}

	return &Reputation{
		CompanyID:   companyID,
		AverageRate: avgRate.Float64,
		ReviewCount: count,
	}, nil
}

func (s *Service) ListByCompanyID(ctx context.Context, companyID string) ([]CompanyReview, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, candidate_id, rating, COALESCE(review, ''), interaction_type, created_at
		 FROM company_reviews
		 WHERE company_id = $1
		 ORDER BY created_at DESC`,
		companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	reviews := []CompanyReview{}
	for rows.Next() {
		var r CompanyReview
		var createdAt time.Time
		if err := rows.Scan(&r.ID, &r.CompanyID, &r.CandidateID, &r.Rating, &r.Review, &r.InteractionType, &createdAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		r.CreatedAt = createdAt.Format(time.RFC3339)
		reviews = append(reviews, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviews: %w", err)
	}

	return reviews, nil
}

func (s *Service) LookupCandidateProfile(ctx context.Context, userID string) (string, error) {
	var profileID uuid.UUID
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM jobseeker_profiles WHERE user_id = $1`, userID,
	).Scan(&profileID)
	if err != nil {
		return "", err
	}
	return profileID.String(), nil
}
