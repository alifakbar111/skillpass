// Package ats implements the full Applicant Tracking System: configurable
// hiring pipelines, per-stage scorecards, interview scheduling, and
// template-based offer letters. On offer acceptance it bridges the candidate
// into the HRIS as an employee, linking their EXISTING login (see onboard.go).
package ats

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
	ErrBadInput  = errors.New("invalid input")
)

// Notifier lets the ATS surface interview/offer events without importing the
// notification package directly (avoids an import cycle; wired in main.go).
type Notifier interface {
	Create(ctx context.Context, userID, notifType, title, body, link string) error
}

type Service struct {
	db       *sql.DB
	notifier Notifier
}

func NewService(db *sql.DB) *Service { return &Service{db: db} }

func (s *Service) SetNotifier(n Notifier) { s.notifier = n }

// ---------- Response shapes (camelCase, swagger-named) ----------

type StageResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StageType string `json:"stageType"`
	SortOrder int    `json:"sortOrder"`
} //@name AtsStage

type PipelineResponse struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	IsDefault bool            `json:"isDefault"`
	CreatedAt string          `json:"createdAt"`
	Stages    []StageResponse `json:"stages,omitempty"`
} //@name AtsPipeline

type CandidateResponse struct {
	ID               string  `json:"id"`
	PipelineID       string  `json:"pipelineId"`
	CurrentStageID   *string `json:"currentStageId,omitempty"`
	CurrentStageName string  `json:"currentStageName,omitempty"`
	ApplicationID    *string `json:"applicationId,omitempty"`
	JobPostingID     *string `json:"jobPostingId,omitempty"`
	JobTitle         string  `json:"jobTitle,omitempty"`
	CandidateName    string  `json:"candidateName"`
	CandidateEmail   string  `json:"candidateEmail"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
} //@name AtsCandidate

// ---------- Pipelines ----------

// defaultStages is the seed sequence for a company's default pipeline.
var defaultStages = []struct {
	Name string
	Type string
}{
	{"Screening", "screening"},
	{"Phone Screen", "phone_screen"},
	{"Technical", "technical"},
	{"HR Interview", "hr_interview"},
	{"Final", "final"},
	{"Offer", "offer"},
	{"Hired", "hired"},
}

// EnsureDefaultPipeline returns the company's default pipeline, creating one
// (with the standard stage sequence) on first use.
func (s *Service) EnsureDefaultPipeline(ctx context.Context, companyID uuid.UUID) (*PipelineResponse, error) {
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM ats_pipelines WHERE company_id=$1 AND is_default=true LIMIT 1`, companyID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		tx, txErr := s.db.BeginTx(ctx, nil)
		if txErr != nil {
			return nil, txErr
		}
		defer tx.Rollback()
		if err = tx.QueryRowContext(ctx,
			`INSERT INTO ats_pipelines (company_id, name, is_default) VALUES ($1,'Default Pipeline',true) RETURNING id`,
			companyID).Scan(&id); err != nil {
			return nil, fmt.Errorf("create default pipeline: %w", err)
		}
		for i, st := range defaultStages {
			if _, err = tx.ExecContext(ctx,
				`INSERT INTO ats_pipeline_stages (pipeline_id, name, stage_type, sort_order) VALUES ($1,$2,$3,$4)`,
				id, st.Name, st.Type, i); err != nil {
				return nil, fmt.Errorf("seed stage: %w", err)
			}
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return s.GetPipeline(ctx, companyID, id)
}

func (s *Service) ListPipelines(ctx context.Context, companyID uuid.UUID) ([]PipelineResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, is_default, created_at FROM ats_pipelines WHERE company_id=$1 ORDER BY is_default DESC, created_at`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PipelineResponse{}
	for rows.Next() {
		var p PipelineResponse
		var created time.Time
		if err := rows.Scan(&p.ID, &p.Name, &p.IsDefault, &created); err != nil {
			return nil, err
		}
		p.CreatedAt = created.Format(time.RFC3339)
		p.Stages = []StageResponse{}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Attach stages per pipeline.
	for i := range out {
		pid, _ := uuid.Parse(out[i].ID)
		stages, err := s.listStages(ctx, pid)
		if err != nil {
			return nil, err
		}
		out[i].Stages = stages
	}
	return out, nil
}

func (s *Service) GetPipeline(ctx context.Context, companyID, pipelineID uuid.UUID) (*PipelineResponse, error) {
	var p PipelineResponse
	var created time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, is_default, created_at FROM ats_pipelines WHERE id=$1 AND company_id=$2`,
		pipelineID, companyID).Scan(&p.ID, &p.Name, &p.IsDefault, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt = created.Format(time.RFC3339)
	stages, err := s.listStages(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	p.Stages = stages
	return &p, nil
}

func (s *Service) listStages(ctx context.Context, pipelineID uuid.UUID) ([]StageResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, stage_type::text, sort_order FROM ats_pipeline_stages WHERE pipeline_id=$1 ORDER BY sort_order`, pipelineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StageResponse{}
	for rows.Next() {
		var st StageResponse
		if err := rows.Scan(&st.ID, &st.Name, &st.StageType, &st.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

type PipelineInput struct {
	Name   string       `json:"name"`
	Stages []StageInput `json:"stages,omitempty"`
}

type StageInput struct {
	Name      string `json:"name"`
	StageType string `json:"stageType"`
}

var validStageType = map[string]bool{
	"screening": true, "phone_screen": true, "technical": true,
	"hr_interview": true, "final": true, "offer": true, "hired": true,
}

func (s *Service) CreatePipeline(ctx context.Context, companyID uuid.UUID, in PipelineInput) (*PipelineResponse, error) {
	if in.Name == "" {
		return nil, ErrBadInput
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var id uuid.UUID
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO ats_pipelines (company_id, name) VALUES ($1,$2) RETURNING id`, companyID, in.Name).Scan(&id); err != nil {
		return nil, err
	}
	for i, st := range in.Stages {
		t := st.StageType
		if !validStageType[t] {
			t = "screening"
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO ats_pipeline_stages (pipeline_id, name, stage_type, sort_order) VALUES ($1,$2,$3,$4)`,
			id, st.Name, t, i); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetPipeline(ctx, companyID, id)
}

// UpdatePipeline replaces a pipeline's name and full stage list (company-scoped).
func (s *Service) UpdatePipeline(ctx context.Context, companyID, pipelineID uuid.UUID, in PipelineInput) (*PipelineResponse, error) {
	if in.Name == "" {
		return nil, ErrBadInput
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `UPDATE ats_pipelines SET name=$1 WHERE id=$2 AND company_id=$3`, in.Name, pipelineID, companyID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	// Replace stages. Candidates referencing removed stages get current_stage_id=NULL via FK ON DELETE SET NULL.
	if _, err := tx.ExecContext(ctx, `DELETE FROM ats_pipeline_stages WHERE pipeline_id=$1`, pipelineID); err != nil {
		return nil, err
	}
	for i, st := range in.Stages {
		t := st.StageType
		if !validStageType[t] {
			t = "screening"
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO ats_pipeline_stages (pipeline_id, name, stage_type, sort_order) VALUES ($1,$2,$3,$4)`,
			pipelineID, st.Name, t, i); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.GetPipeline(ctx, companyID, pipelineID)
}

func (s *Service) DeletePipeline(ctx context.Context, companyID, pipelineID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM ats_pipelines WHERE id=$1 AND company_id=$2 AND is_default=false`, pipelineID, companyID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Candidates ----------

type AddCandidateInput struct {
	CandidateName  string  `json:"candidateName"`
	CandidateEmail string  `json:"candidateEmail"`
	PipelineID     *string `json:"pipelineId,omitempty"`
	ApplicationID  *string `json:"applicationId,omitempty"`
}

// AddCandidate enters a candidate into a pipeline at its first stage. If an
// applicationId is supplied, candidate details + jobseeker/job links are pulled
// from the existing application (the ATS↔applications bridge).
func (s *Service) AddCandidate(ctx context.Context, companyID uuid.UUID, in AddCandidateInput) (*CandidateResponse, error) {
	pipeline, err := s.resolvePipeline(ctx, companyID, in.PipelineID)
	if err != nil {
		return nil, err
	}
	var firstStage *uuid.UUID
	if len(pipeline.Stages) > 0 {
		id, _ := uuid.Parse(pipeline.Stages[0].ID)
		firstStage = &id
	}
	pipelineID, _ := uuid.Parse(pipeline.ID)

	name, emailAddr := in.CandidateName, in.CandidateEmail
	var appID, jobseekerID, jobPostingID *uuid.UUID

	if in.ApplicationID != nil && *in.ApplicationID != "" {
		aID, perr := uuid.Parse(*in.ApplicationID)
		if perr != nil {
			return nil, ErrBadInput
		}
		var jsID, jpID uuid.UUID
		var cName, cEmail string
		err := s.db.QueryRowContext(ctx, `
			SELECT a.jobseeker_id, a.job_posting_id, u.name, u.email
			FROM applications a
			JOIN job_postings jp ON jp.id = a.job_posting_id
			JOIN jobseeker_profiles jsp ON jsp.id = a.jobseeker_id
			JOIN users u ON u.id = jsp.user_id
			WHERE a.id=$1 AND jp.company_id=$2`, aID, companyID).Scan(&jsID, &jpID, &cName, &cEmail)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		appID, jobseekerID, jobPostingID = &aID, &jsID, &jpID
		if name == "" {
			name = cName
		}
		if emailAddr == "" {
			emailAddr = cEmail
		}
	}
	if name == "" || emailAddr == "" {
		return nil, ErrBadInput
	}

	var id uuid.UUID
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO ats_candidates
			(company_id, pipeline_id, current_stage_id, application_id, jobseeker_id, job_posting_id, candidate_name, candidate_email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`,
		companyID, pipelineID, firstStage, appID, jobseekerID, jobPostingID, name, emailAddr).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: candidate already in pipeline", ErrBadInput)
		}
		return nil, err
	}
	return s.GetCandidate(ctx, companyID, id)
}

func (s *Service) resolvePipeline(ctx context.Context, companyID uuid.UUID, pipelineID *string) (*PipelineResponse, error) {
	if pipelineID != nil && *pipelineID != "" {
		pid, err := uuid.Parse(*pipelineID)
		if err != nil {
			return nil, ErrBadInput
		}
		return s.GetPipeline(ctx, companyID, pid)
	}
	return s.EnsureDefaultPipeline(ctx, companyID)
}

func (s *Service) ListCandidates(ctx context.Context, companyID uuid.UUID, pipelineID string) ([]CandidateResponse, error) {
	q := `
		SELECT c.id, c.pipeline_id, c.current_stage_id, COALESCE(st.name,''),
		       c.application_id, c.job_posting_id, COALESCE(jp.title,''),
		       c.candidate_name, c.candidate_email, c.status::text, c.created_at, c.updated_at
		FROM ats_candidates c
		LEFT JOIN ats_pipeline_stages st ON st.id = c.current_stage_id
		LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		WHERE c.company_id=$1`
	args := []any{companyID}
	if pipelineID != "" {
		args = append(args, pipelineID)
		q += fmt.Sprintf(" AND c.pipeline_id=$%d", len(args))
	}
	q += " ORDER BY c.created_at DESC"
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCandidates(rows)
}

func (s *Service) GetCandidate(ctx context.Context, companyID, candidateID uuid.UUID) (*CandidateResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.pipeline_id, c.current_stage_id, COALESCE(st.name,''),
		       c.application_id, c.job_posting_id, COALESCE(jp.title,''),
		       c.candidate_name, c.candidate_email, c.status::text, c.created_at, c.updated_at
		FROM ats_candidates c
		LEFT JOIN ats_pipeline_stages st ON st.id = c.current_stage_id
		LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		WHERE c.company_id=$1 AND c.id=$2`, companyID, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, err := scanCandidates(rows)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNotFound
	}
	return &list[0], nil
}

func scanCandidates(rows *sql.Rows) ([]CandidateResponse, error) {
	out := []CandidateResponse{}
	for rows.Next() {
		var c CandidateResponse
		var stageID, appID, jobPostingID *string
		var created, updated time.Time
		if err := rows.Scan(&c.ID, &c.PipelineID, &stageID, &c.CurrentStageName,
			&appID, &jobPostingID, &c.JobTitle, &c.CandidateName, &c.CandidateEmail,
			&c.Status, &created, &updated); err != nil {
			return nil, err
		}
		c.CurrentStageID = stageID
		c.ApplicationID = appID
		c.JobPostingID = jobPostingID
		c.CreatedAt = created.Format(time.RFC3339)
		c.UpdatedAt = updated.Format(time.RFC3339)
		out = append(out, c)
	}
	return out, rows.Err()
}

// MoveCandidate moves a candidate to a target stage in the same pipeline and
// keeps application/candidate status in sync (offer→n/a; hired handled on
// offer acceptance, rejection is explicit).
func (s *Service) MoveCandidate(ctx context.Context, companyID, candidateID, stageID uuid.UUID) (*CandidateResponse, error) {
	// Ensure the target stage belongs to the candidate's pipeline.
	res, err := s.db.ExecContext(ctx, `
		UPDATE ats_candidates c
		SET current_stage_id=$3, updated_at=now()
		WHERE c.id=$1 AND c.company_id=$2
		  AND EXISTS (SELECT 1 FROM ats_pipeline_stages st WHERE st.id=$3 AND st.pipeline_id=c.pipeline_id)`,
		candidateID, companyID, stageID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.GetCandidate(ctx, companyID, candidateID)
}

// SetCandidateStatus updates a candidate's status (e.g. rejected/withdrawn).
func (s *Service) SetCandidateStatus(ctx context.Context, companyID, candidateID uuid.UUID, status string) (*CandidateResponse, error) {
	switch status {
	case "active", "rejected", "withdrawn", "hired":
	default:
		return nil, ErrBadInput
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE ats_candidates SET status=$3, updated_at=now() WHERE id=$1 AND company_id=$2`,
		candidateID, companyID, status)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.GetCandidate(ctx, companyID, candidateID)
}

// ---------- Scorecards ----------

type ScorecardResponse struct {
	ID             string          `json:"id"`
	CandidateID    string          `json:"candidateId"`
	StageID        *string         `json:"stageId,omitempty"`
	EvaluatorName  string          `json:"evaluatorName"`
	Scores         json.RawMessage `json:"scores,omitempty"`
	OverallRating  *int            `json:"overallRating,omitempty"`
	Recommendation *string         `json:"recommendation,omitempty"`
	Notes          *string         `json:"notes,omitempty"`
	CreatedAt      string          `json:"createdAt"`
} //@name AtsScorecard

type ScorecardInput struct {
	StageID        *string         `json:"stageId,omitempty"`
	Scores         json.RawMessage `json:"scores,omitempty"`
	OverallRating  *int            `json:"overallRating,omitempty"`
	Recommendation *string         `json:"recommendation,omitempty"`
	Notes          *string         `json:"notes,omitempty"`
}

func (s *Service) AddScorecard(ctx context.Context, companyID, candidateID uuid.UUID, evaluatorID uuid.UUID, evaluatorName string, in ScorecardInput) (*ScorecardResponse, error) {
	if err := s.assertCandidate(ctx, companyID, candidateID); err != nil {
		return nil, err
	}
	scores := in.Scores
	if len(scores) == 0 {
		scores = json.RawMessage(`{}`)
	}
	if in.OverallRating != nil && (*in.OverallRating < 1 || *in.OverallRating > 5) {
		return nil, ErrBadInput
	}
	if evaluatorName == "" {
		_ = s.db.QueryRowContext(ctx, `SELECT name FROM users WHERE id=$1`, evaluatorID).Scan(&evaluatorName)
	}
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO ats_scorecards (candidate_id, stage_id, evaluator_id, evaluator_name, scores, overall_rating, recommendation, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		candidateID, in.StageID, evaluatorID, evaluatorName, scores, in.OverallRating, in.Recommendation, in.Notes).Scan(&id)
	if err != nil {
		return nil, err
	}
	list, err := s.ListScorecards(ctx, companyID, candidateID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id.String() {
			return &list[i], nil
		}
	}
	return nil, ErrNotFound
}

func (s *Service) ListScorecards(ctx context.Context, companyID, candidateID uuid.UUID) ([]ScorecardResponse, error) {
	if err := s.assertCandidate(ctx, companyID, candidateID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, candidate_id, stage_id, evaluator_name, scores, overall_rating, recommendation, notes, created_at
		FROM ats_scorecards WHERE candidate_id=$1 ORDER BY created_at DESC`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ScorecardResponse{}
	for rows.Next() {
		var sc ScorecardResponse
		var scores []byte
		var created time.Time
		if err := rows.Scan(&sc.ID, &sc.CandidateID, &sc.StageID, &sc.EvaluatorName, &scores,
			&sc.OverallRating, &sc.Recommendation, &sc.Notes, &created); err != nil {
			return nil, err
		}
		sc.Scores = json.RawMessage(scores)
		sc.CreatedAt = created.Format(time.RFC3339)
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *Service) assertCandidate(ctx context.Context, companyID, candidateID uuid.UUID) error {
	var ok bool
	if err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM ats_candidates WHERE id=$1 AND company_id=$2)`, candidateID, companyID).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint"))
}
