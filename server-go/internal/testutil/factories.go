package testutil

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/models"
)

// bunFor wraps a *sql.DB in a *bun.DB. Factories take *sql.DB (the return
// type of SetupTestDB) and convert internally so test code does not have to
// hold two handles.
func bunFor(db *sql.DB) *bun.DB {
	return bun.NewDB(db, pgdialect.New())
}

func CreateUser(db *sql.DB, email, username, password, name, role string) (uuid.UUID, error) {
	hash, err := lib.HashPassword(password)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash password: %w", err)
	}
	user := &models.User{
		Email:        email,
		Username:     username,
		PasswordHash: hash,
		Name:         name,
		Role:         role,
		CreatedAt:    time.Now(),
	}
	_, err = bunFor(db).NewInsert().Model(user).Returning("id").Exec(context.Background())
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}
	return user.ID, nil
}

func CreateJobseeker(db *sql.DB, email, username, password, name string) (uuid.UUID, uuid.UUID, error) {
	userID, err := CreateUser(db, email, username, password, name, "jobseeker")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	profile := &models.JobseekerProfile{
		ID:     uuid.New(),
		UserID: userID,
		Slug:   username,
	}
	_, err = bunFor(db).NewInsert().Model(profile).Exec(context.Background())
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("insert profile: %w", err)
	}
	return userID, profile.ID, nil
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
	docs := `{"businessRegistration":"reg123","website":"https://example.com","address":"123 Main St","contact":"contact@example.com"}`
	company := &models.Company{
		ID:                 companyID,
		UserID:             userID,
		CompanyName:        companyName,
		Industry:           "Technology",
		VerificationDocs:   &docs,
		VerificationStatus: verificationStatus,
		CreatedAt:          time.Now(),
	}
	bunDB := bunFor(db)
	_, err = bunDB.NewInsert().Model(company).Exec(context.Background())
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("insert company: %w", err)
	}
	if verified {
		_, err = bunDB.NewUpdate().Model(&models.User{}).Set("is_verified = ?", true).Where("id = ?", userID).Exec(context.Background())
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
	job := &models.JobPosting{
		ID:          jobID,
		CompanyID:   companyID,
		Title:       title,
		Description: "Test job description",
		Industry:    industry,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_, err := bunFor(db).NewInsert().Model(job).Exec(context.Background())
	return jobID, err
}

func CreateExperience(db *sql.DB, profileID uuid.UUID, expType, title, org string) (uuid.UUID, error) {
	expID := uuid.New()
	exp := &models.JobExperience{
		ID:           expID,
		ProfileID:    profileID,
		Type:         expType,
		Title:        title,
		Organization: org,
		StartDate:    "2020-01",
		IsCurrent:    true,
	}
	_, err := bunFor(db).NewInsert().Model(exp).Exec(context.Background())
	return expID, err
}

func InsertRefreshToken(db *sql.DB, tokenID uuid.UUID, userID uuid.UUID, tokenStr string, expiresAt time.Time) error {
	hash := sha256.Sum256([]byte(tokenStr))
	tokenHash := hex.EncodeToString(hash[:])
	rt := &models.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	_, err := bunFor(db).NewInsert().Model(rt).Exec(context.Background())
	return err
}

func CreateIndustry(db *sql.DB, name, description string) error {
	industry := &models.IndustryCategory{
		Name:        name,
		Description: &description,
	}
	_, err := bunFor(db).NewInsert().Model(industry).Exec(context.Background())
	return err
}

func CreateTag(db *sql.DB, name, industryID string) error {
	uid, err := uuid.Parse(industryID)
	if err != nil {
		return fmt.Errorf("parse industry id: %w", err)
	}
	tag := &models.Tag{
		Name:               name,
		IndustryCategoryID: &uid,
	}
	_, err = bunFor(db).NewInsert().Model(tag).Exec(context.Background())
	return err
}

func CreateAIEvaluation(db *sql.DB, profileID uuid.UUID, overallScore int) error {
	eval := &models.Evaluation{
		ID:           uuid.New(),
		ProfileID:    profileID,
		OverallScore: int32(overallScore),
		Strengths:    `[{"skill":"Go","score":90,"note":"Strong"}]`,
		Weaknesses:   `[{"skill":"React","score":40,"note":"Weak"}]`,
		Suggestions:  `[{"area":"Frontend","tip":"Learn React"}]`,
		SkillScores:  `[{"skill":"Go","category":"backend","score":90}]`,
		CreatedAt:    time.Now(),
		IsCurrent:    true,
	}
	_, err := bunFor(db).NewInsert().Model(eval).Exec(context.Background())
	return err
}

func CreateApplication(db *sql.DB, profileID, jobPostingID uuid.UUID, status string) (uuid.UUID, error) {
	app := &models.Application{
		ID:           uuid.New(),
		JobseekerID:  profileID,
		JobPostingID: jobPostingID,
		Status:       status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_, err := bunFor(db).NewInsert().Model(app).Exec(context.Background())
	return app.ID, err
}
