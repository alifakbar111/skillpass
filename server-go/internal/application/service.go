package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

// Sentinel errors for error type discrimination.
var (
	ErrJobNotFound     = errors.New("job posting not found")
	ErrJobClosed       = errors.New("job posting is not open for applications")
	ErrDuplicate       = errors.New("already applied to this job")
	ErrInvalidStatus   = errors.New("invalid status")
	ErrAppNotFound     = errors.New("application not found")
	ErrProfileNotFound = errors.New("jobseeker profile not found")
	ErrForbidden       = errors.New("company does not own this application")
)

// Allowed status transitions.
var allowedTransitions = map[string][]string{
	"applied":     {"reviewed", "rejected"},
	"reviewed":    {"interviewed", "rejected"},
	"interviewed": {"offered", "rejected"},
	"offered":     {"rejected"},
	"rejected":    {},
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type ApplicationResult struct {
	ID           string `json:"id"`
	JobseekerID  string `json:"jobseekerId"`
	JobPostingID string `json:"jobPostingId"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	// Joined fields
	JobTitle    string `json:"jobTitle,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
}

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

func (s *Service) Apply(ctx context.Context, jobseekerID, jobPostingID string) (*ApplicationResult, error) {
	// Verify job posting exists and is open
	jobStmt := SELECT(
		gen.JobPostings.ID, gen.JobPostings.Status,
	).FROM(
		gen.JobPostings,
	).WHERE(
		gen.JobPostings.ID.EQ(UUID(uuid.MustParse(jobPostingID))),
	)

	var jobs []struct {
		model.JobPostings
	}
	err := jobStmt.QueryContext(ctx, s.db, &jobs)
	if err != nil {
		return nil, fmt.Errorf("query job posting: %w", err)
	}
	if len(jobs) == 0 {
		return nil, ErrJobNotFound
	}
	job := jobs[0]
	if string(job.Status) != "open" {
		return nil, ErrJobClosed
	}

	// Check for duplicate application using slice target
	dupStmt := SELECT(
		gen.Applications.ID,
	).FROM(
		gen.Applications,
	).WHERE(
		gen.Applications.JobseekerID.EQ(UUID(uuid.MustParse(jobseekerID))).
			AND(gen.Applications.JobPostingID.EQ(UUID(uuid.MustParse(jobPostingID)))),
	).LIMIT(1)

	var dups []model.Applications
	err = dupStmt.QueryContext(ctx, s.db, &dups)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if len(dups) > 0 {
		return nil, ErrDuplicate
	}

	// Insert with RETURNING
	newID := uuid.New()
	insStmt := gen.Applications.INSERT(
		gen.Applications.ID,
		gen.Applications.JobseekerID,
		gen.Applications.JobPostingID,
		gen.Applications.Status,
	).VALUES(
		newID,
		uuid.MustParse(jobseekerID),
		uuid.MustParse(jobPostingID),
		"applied",
	).RETURNING(
		gen.Applications.AllColumns,
	)

	var app model.Applications
	if err := insStmt.QueryContext(ctx, s.db, &app); err != nil {
		return nil, fmt.Errorf("insert application: %w", err)
	}

	return &ApplicationResult{
		ID:           app.ID.String(),
		JobseekerID:  app.JobseekerID.String(),
		JobPostingID: app.JobPostingID.String(),
		Status:       string(app.Status),
		CreatedAt:    app.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    app.UpdatedAt.Format(time.RFC3339),
	}, nil
} 

func (s *Service) ListForJobseeker(ctx context.Context, jobseekerID string) ([]ApplicationResult, error) {
	stmt := SELECT(
		gen.Applications.AllColumns,
		gen.JobPostings.Title,
		gen.Companies.CompanyName,
	).FROM(
		gen.Applications.
			INNER_JOIN(gen.JobPostings, gen.JobPostings.ID.EQ(gen.Applications.JobPostingID)).
			INNER_JOIN(gen.Companies, gen.Companies.ID.EQ(gen.JobPostings.CompanyID)),
	).WHERE(
		gen.Applications.JobseekerID.EQ(UUID(uuid.MustParse(jobseekerID))),
	).ORDER_BY(
		gen.Applications.CreatedAt.DESC(),
	)

	var rows []struct {
		model.Applications
		Title       string
		CompanyName string
	}
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}

	results := make([]ApplicationResult, len(rows))
	for i, r := range rows {
		results[i] = ApplicationResult{
			ID:           r.ID.String(),
			JobseekerID:  r.JobseekerID.String(),
			JobPostingID: r.JobPostingID.String(),
			Status:       string(r.Status),
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    r.UpdatedAt.Format(time.RFC3339),
			JobTitle:     r.Title,
			CompanyName:  r.CompanyName,
		}
	}
	return results, nil
}

func (s *Service) UpdateStatus(ctx context.Context, applicationID, companyID, status string) (*ApplicationResult, error) {
	validStatuses := map[string]bool{
		"applied": true, "reviewed": true, "interviewed": true,
		"offered": true, "rejected": true,
	}
	if !validStatuses[status] {
		return nil, ErrInvalidStatus
	}

	// First get the current application with its job posting's company
	selectStmt := SELECT(
		gen.Applications.AllColumns,
		gen.JobPostings.CompanyID,
	).FROM(
		gen.Applications.
			INNER_JOIN(gen.JobPostings, gen.JobPostings.ID.EQ(gen.Applications.JobPostingID)),
	).WHERE(
		gen.Applications.ID.EQ(UUID(uuid.MustParse(applicationID))),
	).LIMIT(1)

	var currentRows []struct {
		model.Applications
		CompanyID uuid.UUID `alias:"job_postings.company_id"`
	}
	err := selectStmt.QueryContext(ctx, s.db, &currentRows)
	if err != nil {
		return nil, fmt.Errorf("query application: %w", err)
	}
	if len(currentRows) == 0 {
		return nil, ErrAppNotFound
	}
	current := currentRows[0]

	// Verify company ownership
	if current.CompanyID.String() != companyID {
		return nil, ErrForbidden
	}

	// Validate status transition
	fromStatus := string(current.Status)
	if !contains(allowedTransitions[fromStatus], status) {
		return nil, fmt.Errorf("cannot transition from %s to %s", fromStatus, status)
	}

	// Update (trigger handles updated_at)
	var statusExpr StringExpression
	switch status {
	case "applied":
		statusExpr = gen.ApplicationStatusApplied
	case "reviewed":
		statusExpr = gen.ApplicationStatusReviewed
	case "interviewed":
		statusExpr = gen.ApplicationStatusInterviewed
	case "offered":
		statusExpr = gen.ApplicationStatusOffered
	case "rejected":
		statusExpr = gen.ApplicationStatusRejected
	default:
		return nil, ErrInvalidStatus
	}
	updStmt := gen.Applications.UPDATE(
		gen.Applications.Status,
	).SET(
		gen.Applications.Status.SET(statusExpr),
	).WHERE(
		gen.Applications.ID.EQ(UUID(uuid.MustParse(applicationID))),
	).RETURNING(
		gen.Applications.AllColumns,
	)

	var apps []model.Applications
	err = updStmt.QueryContext(ctx, s.db, &apps)
	if err != nil {
		return nil, fmt.Errorf("update application: %w", err)
	}
	if len(apps) == 0 {
		return nil, ErrAppNotFound
	}
	app := apps[0]

	return &ApplicationResult{
		ID:           app.ID.String(),
		JobseekerID:  app.JobseekerID.String(),
		JobPostingID: app.JobPostingID.String(),
		Status:       string(app.Status),
		CreatedAt:    app.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    app.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) LookupJobseekerProfileID(ctx context.Context, userID string) (string, error) {
	var profileID uuid.UUID
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM jobseeker_profiles WHERE user_id = $1`, userID,
	).Scan(&profileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrProfileNotFound
		}
		return "", fmt.Errorf("lookup jobseeker profile: %w", err)
	}
	return profileID.String(), nil
}
