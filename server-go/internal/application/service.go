package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/evaluation"
	"skillpass-server-go/internal/gen"
	"skillpass-server-go/internal/lib"
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
	db          *sql.DB
	bun         bun.IDB
	evalService *evaluation.Service
}

func NewService(db *sql.DB, bun bun.IDB) *Service {
	return &Service{db: db, bun: bun}
}

func (s *Service) SetEvalService(ev *evaluation.Service) {
	s.evalService = ev
}

type ApplicationResult struct {
	ID           string `json:"id"`
	JobseekerID  string `json:"jobseekerId"`
	JobPostingID string `json:"jobPostingId"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	// Joined fields
	JobTitle    string  `json:"jobTitle,omitempty"`
	CompanyName string  `json:"companyName,omitempty"`
	LatestNote  *string `json:"latestNote,omitempty"`
} //@name ApplicationResult

// Message is a note attached to an application (e.g. company → candidate).
type Message struct {
	ID         string `json:"id"`
	SenderName string `json:"senderName"`
	Body       string `json:"body"`
	CreatedAt  string `json:"createdAt"`
} //@name ApplicationMessage

func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}

func (s *Service) Apply(ctx context.Context, jobseekerID, jobPostingID string) (*ApplicationResult, error) {
	jobseekerUUID, err := lib.ParseUUID(jobseekerID)
	if err != nil {
		return nil, fmt.Errorf("invalid jobseeker ID: %w", err)
	}
	jobPostingUUID, err := lib.ParseUUID(jobPostingID)
	if err != nil {
		return nil, fmt.Errorf("invalid job posting ID: %w", err)
	}

	// Verify job posting exists and is open
	jobStmt := SELECT(
		gen.JobPostings.ID, gen.JobPostings.Status,
	).FROM(
		gen.JobPostings,
	).WHERE(
		gen.JobPostings.ID.EQ(UUID(jobPostingUUID)),
	)

	var jobs []struct {
		model.JobPostings
	}
	err = jobStmt.QueryContext(ctx, s.db, &jobs)
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
		gen.Applications.JobseekerID.EQ(UUID(jobseekerUUID)).
			AND(gen.Applications.JobPostingID.EQ(UUID(jobPostingUUID))),
	).LIMIT(1)

	var dups []model.Applications
	err = dupStmt.QueryContext(ctx, s.db, &dups)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if len(dups) > 0 {
		return nil, ErrDuplicate
	}

	// Check if evaluation is expired — auto-trigger if so
	if s.evalService != nil {
		latestEval, err := s.evalService.GetLatest(ctx, jobseekerID)
		if err == nil {
			createdAt, parseErr := time.Parse(time.RFC3339, latestEval.CreatedAt)
			if parseErr == nil && evaluation.IsExpired(createdAt) {
				slog.Info("auto-triggering evaluation for expired evaluation", "profileID", jobseekerID)
				if _, evalErr := s.evalService.Evaluate(ctx, jobseekerID); evalErr != nil {
					slog.Warn("auto-trigger evaluation failed, continuing with application", "error", evalErr)
				}
			}
		} else if errors.Is(err, sql.ErrNoRows) {
			slog.Info("no evaluation found for profile, allowing application without auto-trigger", "profileID", jobseekerID)
		}
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
		jobseekerUUID,
		jobPostingUUID,
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
	jobseekerUUID, err := lib.ParseUUID(jobseekerID)
	if err != nil {
		return nil, fmt.Errorf("invalid jobseeker ID: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		WITH applications_with_company AS (
			SELECT a.id, a.jobseeker_id, a.job_posting_id, a.status, a.created_at, a.updated_at,
			       jp.title, c.company_name
			FROM applications a
			JOIN job_postings jp ON a.job_posting_id = jp.id
			JOIN companies c ON jp.company_id = c.id
			WHERE a.jobseeker_id = $1
		),
		latest_notes AS (
			SELECT DISTINCT ON (application_id) application_id, body
			FROM application_messages
			WHERE application_id IN (SELECT id FROM applications_with_company)
			ORDER BY application_id, created_at DESC
		)
		SELECT awc.id, awc.jobseeker_id, awc.job_posting_id, awc.status, awc.created_at, awc.updated_at,
		       awc.title, awc.company_name, ln.body as latest_note
		FROM applications_with_company awc
		LEFT JOIN latest_notes ln ON awc.id = ln.application_id
		ORDER BY awc.created_at DESC`,
		jobseekerUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()

	var results []ApplicationResult
	for rows.Next() {
		var id, jobseekerIDStr, jobPostingIDStr, status string
		var createdAt, updatedAt time.Time
		var title, companyName string
		var latestNote *string
		if err := rows.Scan(&id, &jobseekerIDStr, &jobPostingIDStr, &status, &createdAt, &updatedAt,
			&title, &companyName, &latestNote); err != nil {
			return nil, fmt.Errorf("scan application row: %w", err)
		}
		results = append(results, ApplicationResult{
			ID:           id,
			JobseekerID:  jobseekerIDStr,
			JobPostingID: jobPostingIDStr,
			Status:       status,
			CreatedAt:    createdAt.Format(time.RFC3339),
			UpdatedAt:    updatedAt.Format(time.RFC3339),
			JobTitle:     title,
			CompanyName:  companyName,
			LatestNote:   latestNote,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list applications rows: %w", err)
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

	applicationUUID, err := lib.ParseUUID(applicationID)
	if err != nil {
		return nil, fmt.Errorf("invalid application ID: %w", err)
	}

	// First get the current application with its job posting's company
	selectStmt := SELECT(
		gen.Applications.AllColumns,
		gen.JobPostings.CompanyID,
	).FROM(
		gen.Applications.
			INNER_JOIN(gen.JobPostings, gen.JobPostings.ID.EQ(gen.Applications.JobPostingID)),
	).WHERE(
		gen.Applications.ID.EQ(UUID(applicationUUID)),
	).LIMIT(1)

	var currentRows []struct {
		model.Applications
		CompanyID uuid.UUID `alias:"job_postings.company_id"`
	}
	err = selectStmt.QueryContext(ctx, s.db, &currentRows)
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
		return nil, ErrInvalidStatus
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
		gen.Applications.ID.EQ(UUID(applicationUUID)),
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

// CompanyApplicationResult extends ApplicationResult with candidate info for company views.
type CompanyApplicationResult struct {
	ID           string `json:"id"`
	JobseekerID  string `json:"jobseekerId"`
	JobPostingID string `json:"jobPostingId"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	JobTitle     string `json:"jobTitle"`
	CandidateName    string `json:"candidateName"`
	CandidateEmail   string `json:"candidateEmail"`
	CandidateSlug    string `json:"candidateSlug"`
	CandidateHeadline string `json:"candidateHeadline,omitempty"`
	LatestNote       *string `json:"latestNote,omitempty"`
} //@name CompanyApplicationResult

func (s *Service) ListForCompany(ctx context.Context, companyID string) ([]CompanyApplicationResult, error) {
	companyUUID, err := lib.ParseUUID(companyID)
	if err != nil {
		return nil, fmt.Errorf("invalid company ID: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		WITH applications_with_candidate AS (
			SELECT a.id, a.jobseeker_id, a.job_posting_id, a.status, a.created_at, a.updated_at,
			       jp.title, u.name, u.email, jp2.slug, jp2.headline
			FROM applications a
			JOIN job_postings jp ON a.job_posting_id = jp.id
			JOIN jobseeker_profiles jp2 ON jp2.id = a.jobseeker_id
			JOIN users u ON u.id = jp2.user_id
			WHERE jp.company_id = $1
		),
		latest_notes AS (
			SELECT DISTINCT ON (application_id) application_id, body
			FROM application_messages
			WHERE application_id IN (SELECT id FROM applications_with_candidate)
			ORDER BY application_id, created_at DESC
		)
		SELECT awc.id, awc.jobseeker_id, awc.job_posting_id, awc.status, awc.created_at, awc.updated_at,
		       awc.title, awc.name, awc.email, awc.slug, awc.headline, ln.body as latest_note
		FROM applications_with_candidate awc
		LEFT JOIN latest_notes ln ON awc.id = ln.application_id
		ORDER BY awc.created_at DESC`,
		companyUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("list company applications: %w", err)
	}
	defer rows.Close()

	var results []CompanyApplicationResult
	for rows.Next() {
		var id, jobseekerIDStr, jobPostingIDStr, status string
		var createdAt, updatedAt time.Time
		var title, name, email, slug string
		var headline *string
		var latestNote *string
		if err := rows.Scan(&id, &jobseekerIDStr, &jobPostingIDStr, &status, &createdAt, &updatedAt,
			&title, &name, &email, &slug, &headline, &latestNote); err != nil {
			return nil, fmt.Errorf("scan company application row: %w", err)
		}
		headlineStr := ""
		if headline != nil {
			headlineStr = *headline
		}
		results = append(results, CompanyApplicationResult{
			ID:                id,
			JobseekerID:       jobseekerIDStr,
			JobPostingID:      jobPostingIDStr,
			Status:            status,
			CreatedAt:         createdAt.Format(time.RFC3339),
			UpdatedAt:         updatedAt.Format(time.RFC3339),
			JobTitle:          title,
			CandidateName:     name,
			CandidateEmail:    email,
			CandidateSlug:     slug,
			CandidateHeadline: headlineStr,
			LatestNote:        latestNote,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list company applications rows: %w", err)
	}
	return results, nil
}

// verifyCompanyOwnsApplication returns nil if the application belongs to a job
// owned by the given company, ErrAppNotFound if missing, ErrForbidden otherwise.
func (s *Service) verifyCompanyOwnsApplication(ctx context.Context, applicationID, companyID string) error {
	var ownerCompanyID uuid.UUID
	err := s.db.QueryRowContext(ctx,
		`SELECT j.company_id
		 FROM applications a
		 JOIN job_postings j ON j.id = a.job_posting_id
		 WHERE a.id = $1`,
		applicationID,
	).Scan(&ownerCompanyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAppNotFound
		}
		return fmt.Errorf("verify application owner: %w", err)
	}
	if ownerCompanyID.String() != companyID {
		return ErrForbidden
	}
	return nil
}

// AddMessage attaches a note to an application. The company must own the application.
func (s *Service) AddMessage(ctx context.Context, applicationID, companyID, senderUserID, body string) (*Message, error) {
	if err := s.verifyCompanyOwnsApplication(ctx, applicationID, companyID); err != nil {
		return nil, err
	}

	var id uuid.UUID
	var createdAt time.Time
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO application_messages (application_id, sender_user_id, message_type, body)
		 VALUES ($1, $2, 'note', $3)
		 RETURNING id, created_at`,
		applicationID, senderUserID, body,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	var senderName string
	_ = s.db.QueryRowContext(ctx, `SELECT name FROM users WHERE id = $1`, senderUserID).Scan(&senderName)

	return &Message{
		ID:         id.String(),
		SenderName: senderName,
		Body:       body,
		CreatedAt:  createdAt.Format(time.RFC3339),
	}, nil
}

// ListMessages returns the message thread for an application owned by the company.
func (s *Service) ListMessages(ctx context.Context, applicationID, companyID string) ([]Message, error) {
	if err := s.verifyCompanyOwnsApplication(ctx, applicationID, companyID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT m.id, u.name, m.body, m.created_at
		 FROM application_messages m
		 JOIN users u ON u.id = m.sender_user_id
		 WHERE m.application_id = $1
		 ORDER BY m.created_at ASC`,
		applicationID,
	)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	messages := []Message{}
	for rows.Next() {
		var m Message
		var id uuid.UUID
		var createdAt time.Time
		if err := rows.Scan(&id, &m.SenderName, &m.Body, &createdAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		m.ID = id.String()
		m.CreatedAt = createdAt.Format(time.RFC3339)
		messages = append(messages, m)
	}
	return messages, rows.Err()
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
