package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/lib/pq"

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
	jobseekerUUID, err := uuid.Parse(jobseekerID)
	if err != nil {
		return nil, fmt.Errorf("invalid jobseeker ID: %w", err)
	}
	jobPostingUUID, err := uuid.Parse(jobPostingID)
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
	jobseekerUUID, err := uuid.Parse(jobseekerID)
	if err != nil {
		return nil, fmt.Errorf("invalid jobseeker ID: %w", err)
	}

	stmt := SELECT(
		gen.Applications.AllColumns,
		gen.JobPostings.Title,
		gen.Companies.CompanyName,
	).FROM(
		gen.Applications.
			INNER_JOIN(gen.JobPostings, gen.JobPostings.ID.EQ(gen.Applications.JobPostingID)).
			INNER_JOIN(gen.Companies, gen.Companies.ID.EQ(gen.JobPostings.CompanyID)),
	).WHERE(
		gen.Applications.JobseekerID.EQ(UUID(jobseekerUUID)),
	).ORDER_BY(
		gen.Applications.CreatedAt.DESC(),
	)

	var rows []struct {
		model.Applications
		Title       string `alias:"job_postings.title"`
		CompanyName string `alias:"companies.company_name"`
	}
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}

	results := make([]ApplicationResult, len(rows))
	ids := make([]string, len(rows))
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
		ids[i] = results[i].ID
	}

	notes, err := s.latestNotesFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range results {
		if note, ok := notes[results[i].ID]; ok {
			n := note
			results[i].LatestNote = &n
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

	applicationUUID, err := uuid.Parse(applicationID)
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
	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return nil, fmt.Errorf("invalid company ID: %w", err)
	}

	stmt := SELECT(
		gen.Applications.AllColumns,
		gen.JobPostings.Title,
		gen.Users.Name,
		gen.Users.Email,
		gen.JobseekerProfiles.Slug,
		gen.JobseekerProfiles.Headline,
	).FROM(
		gen.Applications.
			INNER_JOIN(gen.JobPostings, gen.JobPostings.ID.EQ(gen.Applications.JobPostingID)).
			INNER_JOIN(gen.JobseekerProfiles, gen.JobseekerProfiles.ID.EQ(gen.Applications.JobseekerID)).
			INNER_JOIN(gen.Users, gen.Users.ID.EQ(gen.JobseekerProfiles.UserID)),
	).WHERE(
		gen.JobPostings.CompanyID.EQ(UUID(companyUUID)),
	).ORDER_BY(
		gen.Applications.CreatedAt.DESC(),
	)

	var rows []struct {
		model.Applications
		Title    string  `alias:"job_postings.title"`
		Name     string  `alias:"users.name"`
		Email    string  `alias:"users.email"`
		Slug     string  `alias:"jobseeker_profiles.slug"`
		Headline *string `alias:"jobseeker_profiles.headline"`
	}
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, fmt.Errorf("list company applications: %w", err)
	}

	results := make([]CompanyApplicationResult, len(rows))
	ids := make([]string, len(rows))
	for i, r := range rows {
		headline := ""
		if r.Headline != nil {
			headline = *r.Headline
		}
		results[i] = CompanyApplicationResult{
			ID:                r.ID.String(),
			JobseekerID:       r.JobseekerID.String(),
			JobPostingID:      r.JobPostingID.String(),
			Status:            string(r.Status),
			CreatedAt:         r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:         r.UpdatedAt.Format(time.RFC3339),
			JobTitle:          r.Title,
			CandidateName:     r.Name,
			CandidateEmail:    r.Email,
			CandidateSlug:     r.Slug,
			CandidateHeadline: headline,
		}
		ids[i] = results[i].ID
	}

	notes, err := s.latestNotesFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range results {
		if note, ok := notes[results[i].ID]; ok {
			n := note
			results[i].LatestNote = &n
		}
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

// latestNotesFor returns a map of applicationID → most recent note body.
func (s *Service) latestNotesFor(ctx context.Context, applicationIDs []string) (map[string]string, error) {
	result := map[string]string{}
	if len(applicationIDs) == 0 {
		return result, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT ON (application_id) application_id, body
		 FROM application_messages
		 WHERE application_id = ANY($1::uuid[])
		 ORDER BY application_id, created_at DESC`,
		pq.Array(applicationIDs),
	)
	if err != nil {
		return nil, fmt.Errorf("query latest notes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var appID uuid.UUID
		var body string
		if err := rows.Scan(&appID, &body); err != nil {
			return nil, fmt.Errorf("scan latest note: %w", err)
		}
		result[appID.String()] = body
	}
	return result, rows.Err()
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
