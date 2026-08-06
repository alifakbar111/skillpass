package ats

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type InterviewResponse struct {
	ID           string  `json:"id"`
	CandidateID  string  `json:"candidateId"`
	StageID      *string `json:"stageId,omitempty"`
	ScheduledAt  string  `json:"scheduledAt"`
	Mode         string  `json:"mode"`
	Location     *string `json:"location,omitempty"`
	MeetingLink  *string `json:"meetingLink,omitempty"`
	Interviewer  *string `json:"interviewer,omitempty"`
	Notes        *string `json:"notes,omitempty"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
} //@name AtsInterview

type InterviewInput struct {
	StageID     *string `json:"stageId,omitempty"`
	ScheduledAt string  `json:"scheduledAt"`
	Mode        string  `json:"mode"`
	Location    string  `json:"location"`
	MeetingLink string  `json:"meetingLink"`
	Interviewer string  `json:"interviewer"`
	Notes       string  `json:"notes"`
}

// ScheduleInterview books an interview for a candidate and notifies the linked
// jobseeker (if the candidate came from an application).
func (s *Service) ScheduleInterview(ctx context.Context, companyID, candidateID, createdBy uuid.UUID, in InterviewInput) (*InterviewResponse, error) {
	if err := s.assertCandidate(ctx, companyID, candidateID); err != nil {
		return nil, err
	}
	when, err := time.Parse(time.RFC3339, in.ScheduledAt)
	if err != nil {
		return nil, ErrBadInput
	}
	mode := in.Mode
	if mode != "online" {
		mode = "onsite"
	}
	var id uuid.UUID
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO ats_interviews (candidate_id, stage_id, scheduled_at, mode, location, meeting_link, interviewer, notes, created_by)
		VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),$9)
		RETURNING id`,
		candidateID, in.StageID, when, mode, in.Location, in.MeetingLink, in.Interviewer, in.Notes, createdBy).Scan(&id)
	if err != nil {
		return nil, err
	}

	s.notifyCandidateUser(ctx, candidateID, "interview",
		"Interview scheduled",
		"An interview has been scheduled for "+when.Format("Mon, 02 Jan 2006 15:04")+".")

	return s.getInterview(ctx, id)
}

func (s *Service) ListInterviews(ctx context.Context, companyID, candidateID uuid.UUID) ([]InterviewResponse, error) {
	if err := s.assertCandidate(ctx, companyID, candidateID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, candidate_id, stage_id, scheduled_at, mode, location, meeting_link, interviewer, notes, status::text, created_at
		FROM ats_interviews WHERE candidate_id=$1 ORDER BY scheduled_at DESC`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []InterviewResponse{}
	for rows.Next() {
		iv, err := scanInterview(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *iv)
	}
	return out, rows.Err()
}

// UpdateInterviewStatus marks an interview completed/cancelled/no_show.
func (s *Service) UpdateInterviewStatus(ctx context.Context, companyID, interviewID uuid.UUID, status string) (*InterviewResponse, error) {
	switch status {
	case "scheduled", "completed", "cancelled", "no_show":
	default:
		return nil, ErrBadInput
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE ats_interviews iv SET status=$3
		WHERE iv.id=$1 AND EXISTS (
			SELECT 1 FROM ats_candidates c WHERE c.id = iv.candidate_id AND c.company_id=$2)`,
		interviewID, companyID, status)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return s.getInterview(ctx, interviewID)
}

func (s *Service) getInterview(ctx context.Context, id uuid.UUID) (*InterviewResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, candidate_id, stage_id, scheduled_at, mode, location, meeting_link, interviewer, notes, status::text, created_at
		FROM ats_interviews WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrNotFound
	}
	return scanInterview(rows)
}

func scanInterview(rows *sql.Rows) (*InterviewResponse, error) {
	var iv InterviewResponse
	var scheduled, created time.Time
	if err := rows.Scan(&iv.ID, &iv.CandidateID, &iv.StageID, &scheduled, &iv.Mode,
		&iv.Location, &iv.MeetingLink, &iv.Interviewer, &iv.Notes, &iv.Status, &created); err != nil {
		return nil, err
	}
	iv.ScheduledAt = scheduled.Format(time.RFC3339)
	iv.CreatedAt = created.Format(time.RFC3339)
	return &iv, nil
}

// notifyCandidateUser sends an in-app notification to the jobseeker's user
// account if this candidate is linked to one. Best-effort.
func (s *Service) notifyCandidateUser(ctx context.Context, candidateID uuid.UUID, notifType, title, body string) {
	if s.notifier == nil {
		return
	}
	var userID sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT jsp.user_id
		FROM ats_candidates c
		JOIN jobseeker_profiles jsp ON jsp.id = c.jobseeker_id
		WHERE c.id=$1`, candidateID).Scan(&userID)
	if err != nil || !userID.Valid {
		return
	}
	_ = s.notifier.Create(ctx, userID.String, notifType, title, body, "/applications")
}
