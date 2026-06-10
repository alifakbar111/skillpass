package analytics

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// Service computes hiring analytics using raw SQL aggregates.
type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// StatusCount is a single status bucket.
type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// JobFunnel summarizes one job posting's application pipeline.
type JobFunnel struct {
	JobPostingID string        `json:"jobPostingId"`
	Title        string        `json:"title"`
	Status       string        `json:"status"`
	Total        int           `json:"total"`
	ByStatus     []StatusCount `json:"byStatus"`
}

// CompanyAnalytics is the company dashboard payload.
type CompanyAnalytics struct {
	TotalJobs            int           `json:"totalJobs"`
	OpenJobs             int           `json:"openJobs"`
	TotalApplications    int           `json:"totalApplications"`
	ApplicationsByStatus []StatusCount `json:"applicationsByStatus"`
	AvgDaysToDecision    *float64      `json:"avgDaysToDecision"`
	Jobs                 []JobFunnel   `json:"jobs"`
}

// JobseekerAnalytics is the jobseeker stats payload.
type JobseekerAnalytics struct {
	TotalApplications    int           `json:"totalApplications"`
	ApplicationsByStatus []StatusCount `json:"applicationsByStatus"`
	PassportViews        int           `json:"passportViews"`
	ResponseRate         *float64      `json:"responseRate"`
}

func (s *Service) ForCompany(ctx context.Context, companyID string) (*CompanyAnalytics, error) {
	result := &CompanyAnalytics{
		ApplicationsByStatus: []StatusCount{},
		Jobs:                 []JobFunnel{},
	}

	// Job counts
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'open')
		 FROM job_postings WHERE company_id = $1`,
		companyID,
	).Scan(&result.TotalJobs, &result.OpenJobs)
	if err != nil {
		return nil, fmt.Errorf("job counts: %w", err)
	}

	// Applications by status (company-wide)
	rows, err := s.db.QueryContext(ctx,
		`SELECT a.status::text, COUNT(*)
		 FROM applications a
		 JOIN job_postings j ON j.id = a.job_posting_id
		 WHERE j.company_id = $1
		 GROUP BY a.status`,
		companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("status counts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, fmt.Errorf("scan status count: %w", err)
		}
		result.ApplicationsByStatus = append(result.ApplicationsByStatus, sc)
		result.TotalApplications += sc.Count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Average days from application to a terminal decision (offered/rejected)
	var avgDays sql.NullFloat64
	err = s.db.QueryRowContext(ctx,
		`SELECT AVG(EXTRACT(EPOCH FROM (a.updated_at - a.created_at)) / 86400.0)
		 FROM applications a
		 JOIN job_postings j ON j.id = a.job_posting_id
		 WHERE j.company_id = $1 AND a.status IN ('offered', 'rejected')`,
		companyID,
	).Scan(&avgDays)
	if err != nil {
		return nil, fmt.Errorf("avg decision time: %w", err)
	}
	if avgDays.Valid {
		v := avgDays.Float64
		result.AvgDaysToDecision = &v
	}

	// Per-job funnels
	jobRows, err := s.db.QueryContext(ctx,
		`SELECT j.id, j.title, j.status::text, a.status::text, COUNT(a.id)
		 FROM job_postings j
		 LEFT JOIN applications a ON a.job_posting_id = j.id
		 WHERE j.company_id = $1
		 GROUP BY j.id, j.title, j.status, a.status
		 ORDER BY j.created_at DESC`,
		companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("job funnels: %w", err)
	}
	defer jobRows.Close()

	funnelIndex := map[string]int{}
	for jobRows.Next() {
		var jobID uuid.UUID
		var title, jobStatus string
		var appStatus sql.NullString
		var count int
		if err := jobRows.Scan(&jobID, &title, &jobStatus, &appStatus, &count); err != nil {
			return nil, fmt.Errorf("scan job funnel: %w", err)
		}
		id := jobID.String()
		idx, ok := funnelIndex[id]
		if !ok {
			result.Jobs = append(result.Jobs, JobFunnel{
				JobPostingID: id,
				Title:        title,
				Status:       jobStatus,
				ByStatus:     []StatusCount{},
			})
			idx = len(result.Jobs) - 1
			funnelIndex[id] = idx
		}
		if appStatus.Valid {
			result.Jobs[idx].ByStatus = append(result.Jobs[idx].ByStatus, StatusCount{Status: appStatus.String, Count: count})
			result.Jobs[idx].Total += count
		}
	}
	if err := jobRows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) ForJobseeker(ctx context.Context, profileID string) (*JobseekerAnalytics, error) {
	result := &JobseekerAnalytics{
		ApplicationsByStatus: []StatusCount{},
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT status::text, COUNT(*)
		 FROM applications
		 WHERE jobseeker_id = $1
		 GROUP BY status`,
		profileID,
	)
	if err != nil {
		return nil, fmt.Errorf("status counts: %w", err)
	}
	defer rows.Close()

	responded := 0
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, fmt.Errorf("scan status count: %w", err)
		}
		result.ApplicationsByStatus = append(result.ApplicationsByStatus, sc)
		result.TotalApplications += sc.Count
		if sc.Status != "applied" {
			responded += sc.Count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if result.TotalApplications > 0 {
		rate := float64(responded) / float64(result.TotalApplications) * 100
		result.ResponseRate = &rate
	}

	if err := s.db.QueryRowContext(ctx,
		`SELECT view_count FROM jobseeker_profiles WHERE id = $1`, profileID,
	).Scan(&result.PassportViews); err != nil {
		return nil, fmt.Errorf("passport views: %w", err)
	}

	return result, nil
}
