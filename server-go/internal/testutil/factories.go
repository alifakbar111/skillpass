package testutil

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/models"
)

func CreateUser(db bun.IDB, email, username, password, name, role string) (uuid.UUID, error) {
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
	_, err = db.NewInsert().Model(user).Returning("id").Exec(context.Background())
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert user: %w", err)
	}
	return user.ID, nil
}

func CreateJobseeker(db bun.IDB, email, username, password, name string) (uuid.UUID, uuid.UUID, error) {
	userID, err := CreateUser(db, email, username, password, name, "jobseeker")
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	profile := &models.JobseekerProfile{
		ID:     uuid.New(),
		UserID: userID,
		Slug:   username,
	}
	_, err = db.NewInsert().Model(profile).Exec(context.Background())
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("insert profile: %w", err)
	}
	return userID, profile.ID, nil
}

func CreateCompanyUser(db bun.IDB, email, username, password, companyName string, verified bool) (uuid.UUID, uuid.UUID, error) {
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
	_, err = db.NewInsert().Model(company).Exec(context.Background())
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("insert company: %w", err)
	}
	if verified {
		_, err = db.NewUpdate().Model(&models.User{}).Set("is_verified = ?", true).Where("id = ?", userID).Exec(context.Background())
		if err != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("mark verified: %w", err)
		}
	}
	return userID, companyID, nil
}

func CreateAdmin(db bun.IDB, email, username, password string) (uuid.UUID, error) {
	return CreateUser(db, email, username, password, "Admin", "admin")
}

func CreateJob(db bun.IDB, companyID uuid.UUID, title, industry string, open bool) (uuid.UUID, error) {
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
	_, err := db.NewInsert().Model(job).Exec(context.Background())
	return jobID, err
}

func CreateExperience(db bun.IDB, profileID uuid.UUID, expType, title, org string) (uuid.UUID, error) {
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
	_, err := db.NewInsert().Model(exp).Exec(context.Background())
	return expID, err
}

func InsertRefreshToken(db bun.IDB, tokenID uuid.UUID, userID uuid.UUID, tokenStr string, expiresAt time.Time) error {
	hash := sha256.Sum256([]byte(tokenStr))
	tokenHash := hex.EncodeToString(hash[:])
	rt := &models.RefreshToken{
		ID:        tokenID,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(rt).Exec(context.Background())
	return err
}

func CreateIndustry(db bun.IDB, name, description string) error {
	industry := &models.IndustryCategory{
		Name:        name,
		Description: &description,
	}
	_, err := db.NewInsert().Model(industry).Exec(context.Background())
	return err
}

func CreateTag(db bun.IDB, name, industryID string) error {
	uid, err := uuid.Parse(industryID)
	if err != nil {
		return fmt.Errorf("parse industry id: %w", err)
	}
	tag := &models.Tag{
		Name:               name,
		IndustryCategoryID: &uid,
	}
	_, err = db.NewInsert().Model(tag).Exec(context.Background())
	return err
}

func CreateAIEvaluation(db bun.IDB, profileID uuid.UUID, overallScore int) error {
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
	_, err := db.NewInsert().Model(eval).Exec(context.Background())
	return err
}

func CreateApplication(db bun.IDB, profileID, jobPostingID uuid.UUID, status string) (uuid.UUID, error) {
	app := &models.Application{
		ID:           uuid.New(),
		JobseekerID:  profileID,
		JobPostingID: jobPostingID,
		Status:       status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_, err := db.NewInsert().Model(app).Exec(context.Background())
	return app.ID, err
}
