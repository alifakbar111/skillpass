package testutil

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/internal/gen"
	"skillpass-server-go/internal/lib"
)

func CreateUser(db *sql.DB, email, username, password, name, role string) (uuid.UUID, error) {
	hash, err := lib.HashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}
	var id uuid.UUID
	err = db.QueryRowContext(context.Background(),
		`INSERT INTO users (email, username, password_hash, name, role) 
		 VALUES ($1, $2, $3, $4, $5::role) RETURNING id`,
		email, username, hash, name, role,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func CreateJobseeker(db *sql.DB, email, username, password, name string) (uuid.UUID, uuid.UUID, error) {
	userID, err := CreateUser(db, email, username, password, name, "jobseeker")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	profileID := uuid.New()
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO jobseeker_profiles (id, user_id, slug) VALUES ($1, $2, $3)`,
		profileID, userID, username,
	)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("insert profile: %w", err)
	}
	return userID, profileID, nil
}

func CreateCompanyUser(db *sql.DB, email, username, password, companyName string, verified bool) (uuid.UUID, uuid.UUID, error) {
	userID, err := CreateUser(db, email, username, password, companyName, "company")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	companyID := uuid.New()
	verificationStatus := "verified"
	if !verified {
		verificationStatus = "pending"
	}
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO companies (id, user_id, company_name, industry, verification_docs, verification_status)
		 VALUES ($1, $2, $3, $4, $5::jsonb, $6::verification_status)`,
		companyID, userID, companyName, "Technology",
		`{"businessRegistration":"reg123","website":"https://example.com","address":"123 Main St","contact":"contact@example.com"}`,
		verificationStatus,
	)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("insert company: %w", err)
	}
	if verified {
		_, err = gen.Users.UPDATE().
			SET(gen.Users.IsVerified.SET(postgres.Bool(true))).
			WHERE(gen.Users.ID.EQ(postgres.UUID(userID))).
			ExecContext(context.Background(), db)
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("mark verified: %w", err)
		}
	}
	return userID, companyID, nil
}

func CreateAdmin(db *sql.DB, email, username, password string) (uuid.UUID, error) {
	return CreateUser(db, email, username, password, "Admin", "admin")
}

func CreateJob(db *sql.DB, companyID uuid.UUID, title, industry string, open bool) (uuid.UUID, error) {
	jobID := uuid.New()
	status := "open"
	if !open {
		status = "closed"
	}
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO job_postings (id, company_id, title, description, industry, status)
		 VALUES ($1, $2, $3, $4, $5, $6::job_status)`,
		jobID, companyID, title, "Test job description", industry, status,
	)
	return jobID, err
}

func CreateExperience(db *sql.DB, profileID uuid.UUID, expType, title, org string) (uuid.UUID, error) {
	expID := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO job_experiences (id, profile_id, type, title, organization, start_date, is_current)
		 VALUES ($1, $2, $3::experience_type, $4, $5, $6, $7)`,
		expID, profileID, expType, title, org, "2020-01", true,
	)
	return expID, err
}

func InsertRefreshToken(db *sql.DB, tokenID uuid.UUID, userID uuid.UUID, tokenStr string, expiresAt time.Time) error {
	hash := sha256.Sum256([]byte(tokenStr))
	tokenHash := hex.EncodeToString(hash[:])
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		tokenID, userID, tokenHash, expiresAt,
	)
	return err
}

func CreateIndustry(db *sql.DB, name, description string) error {
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO industry_categories (name, description) VALUES ($1, $2)`,
		name, description,
	)
	return err
}

func CreateTag(db *sql.DB, name, industryID string) error {
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO tags (name, industry_category_id) VALUES ($1, $2)`,
		name, industryID,
	)
	return err
}

func CreateAIEvaluation(db *sql.DB, profileID uuid.UUID, overallScore int) error {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO ai_evaluations (id, profile_id, overall_score, strengths, weaknesses, suggestions, skill_scores)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, profileID, overallScore,
		`[{"skill":"Go","score":90,"note":"Strong"}]`,
		`[{"skill":"React","score":40,"note":"Weak"}]`,
		`[{"area":"Frontend","tip":"Learn React"}]`,
		`[{"skill":"Go","category":"backend","score":90}]`,
	)
	return err
}

func CreateApplication(db *sql.DB, profileID, jobPostingID uuid.UUID, status string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO applications (id, jobseeker_id, job_posting_id, status)
		 VALUES ($1, $2, $3, $4::application_status)`,
		id, profileID, jobPostingID, status,
	)
	return id, err
}
