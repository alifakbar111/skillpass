package leave

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type LeaveType struct {
	ID                 uuid.UUID `json:"id"`
	CompanyID          uuid.UUID `json:"companyId"`
	Name               string    `json:"name"`
	Code               string    `json:"code"`
	DefaultDaysPerYear int       `json:"defaultDaysPerYear"`
	IsPaid             bool      `json:"isPaid"`
	RequiresAttachment bool      `json:"requiresAttachment"`
	IsActive           bool      `json:"isActive"`
	CreatedAt          time.Time `json:"createdAt"`
}

type LeaveBalance struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employeeId"`
	LeaveTypeID   uuid.UUID `json:"leaveTypeId"`
	LeaveTypeName string    `json:"leaveTypeName,omitempty"`
	Year          int       `json:"year"`
	TotalDays     int       `json:"totalDays"`
	UsedDays      int       `json:"usedDays"`
	CarryOverDays int       `json:"carryOverDays"`
	Remaining     int       `json:"remaining"`
	CreatedAt     time.Time `json:"createdAt"`
}

type LeaveRequest struct {
	ID              uuid.UUID  `json:"id"`
	CompanyID       uuid.UUID  `json:"companyId"`
	EmployeeID      uuid.UUID  `json:"employeeId"`
	EmployeeName    string     `json:"employeeName,omitempty"`
	LeaveTypeID     uuid.UUID  `json:"leaveTypeId"`
	LeaveTypeName   string     `json:"leaveTypeName,omitempty"`
	StartDate       string     `json:"startDate"`
	EndDate         string     `json:"endDate"`
	TotalDays       int        `json:"totalDays"`
	Reason          string     `json:"reason"`
	AttachmentURL   *string    `json:"attachmentUrl"`
	Status          string     `json:"status"`
	ReviewerID      *uuid.UUID `json:"reviewerId"`
	ReviewerComment *string    `json:"reviewerComment"`
	ReviewedAt      *time.Time `json:"reviewedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// ── Leave Types ──

func (s *Service) ListTypes(ctx context.Context, companyID uuid.UUID) ([]LeaveType, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, code, default_days_per_year, is_paid, requires_attachment, is_active, created_at
		 FROM leave_types WHERE company_id = $1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []LeaveType
	for rows.Next() {
		var t LeaveType
		if err := rows.Scan(&t.ID, &t.CompanyID, &t.Name, &t.Code, &t.DefaultDaysPerYear,
			&t.IsPaid, &t.RequiresAttachment, &t.IsActive, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (s *Service) CreateType(ctx context.Context, companyID uuid.UUID, t *LeaveType) error {
	return s.db.QueryRowContext(ctx,
		`INSERT INTO leave_types (company_id, name, code, default_days_per_year, is_paid, requires_attachment)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, is_active, created_at`,
		companyID, t.Name, t.Code, t.DefaultDaysPerYear, t.IsPaid, t.RequiresAttachment,
	).Scan(&t.ID, &t.IsActive, &t.CreatedAt)
}

func (s *Service) UpdateType(ctx context.Context, companyID, id uuid.UUID, t *LeaveType) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE leave_types SET name=$1, code=$2, default_days_per_year=$3, is_paid=$4, requires_attachment=$5, is_active=$6
		 WHERE id=$7 AND company_id=$8`,
		t.Name, t.Code, t.DefaultDaysPerYear, t.IsPaid, t.RequiresAttachment, t.IsActive, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Service) DeleteType(ctx context.Context, companyID, id uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM leave_types WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ── Leave Balances ──

func (s *Service) GetBalances(ctx context.Context, companyID, employeeID uuid.UUID, year int) ([]LeaveBalance, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT lb.id, lb.employee_id, lb.leave_type_id, lt.name, lb.year,
		        lb.total_days, lb.used_days, lb.carry_over_days,
		        (lb.total_days + lb.carry_over_days - lb.used_days) as remaining,
		        lb.created_at
		 FROM leave_balances lb
		 JOIN leave_types lt ON lt.id = lb.leave_type_id
		 JOIN employees e ON e.id = lb.employee_id
		 WHERE e.company_id = $1 AND lb.employee_id = $2 AND lb.year = $3
		 ORDER BY lt.name`, companyID, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []LeaveBalance
	for rows.Next() {
		var b LeaveBalance
		if err := rows.Scan(&b.ID, &b.EmployeeID, &b.LeaveTypeID, &b.LeaveTypeName, &b.Year,
			&b.TotalDays, &b.UsedDays, &b.CarryOverDays, &b.Remaining, &b.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}

func (s *Service) InitBalances(ctx context.Context, companyID, employeeID uuid.UUID, year int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO leave_balances (employee_id, leave_type_id, year, total_days)
		 SELECT $1, id, $2, default_days_per_year
		 FROM leave_types WHERE company_id = $3 AND is_active = true
		 ON CONFLICT (employee_id, leave_type_id, year) DO NOTHING`,
		employeeID, year, companyID)
	return err
}

// ── Leave Requests ──

func (s *Service) CreateRequest(ctx context.Context, companyID, employeeID uuid.UUID, req *LeaveRequest) error {
	var remaining int
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(total_days + carry_over_days - used_days, 0)
		 FROM leave_balances
		 WHERE employee_id = $1 AND leave_type_id = $2 AND year = EXTRACT(YEAR FROM $3::date)`,
		employeeID, req.LeaveTypeID, req.StartDate).Scan(&remaining)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no leave balance found — run balance init first")
		}
		return err
	}
	if remaining < req.TotalDays {
		return fmt.Errorf("insufficient leave balance: %d remaining, %d requested", remaining, req.TotalDays)
	}

	return s.db.QueryRowContext(ctx,
		`INSERT INTO leave_requests (company_id, employee_id, leave_type_id, start_date, end_date, total_days, reason, attachment_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, status, created_at`,
		companyID, employeeID, req.LeaveTypeID, req.StartDate, req.EndDate, req.TotalDays, req.Reason, req.AttachmentURL,
	).Scan(&req.ID, &req.Status, &req.CreatedAt)
}

func (s *Service) ListRequests(ctx context.Context, companyID uuid.UUID, status string) ([]LeaveRequest, error) {
	query := `SELECT lr.id, lr.company_id, lr.employee_id,
	                 COALESCE(e.first_name||' '||e.last_name,'') as employee_name,
	                 lr.leave_type_id, lt.name as leave_type_name,
	                 lr.start_date::text, lr.end_date::text, lr.total_days,
	                 lr.reason, lr.attachment_url, lr.status,
	                 lr.reviewer_id, lr.reviewer_comment, lr.reviewed_at, lr.created_at
	          FROM leave_requests lr
	          JOIN employees e ON e.id = lr.employee_id
	          JOIN leave_types lt ON lt.id = lr.leave_type_id
	          WHERE lr.company_id = $1`
	args := []any{companyID}
	if status != "" {
		query += " AND lr.status = $2"
		args = append(args, status)
	}
	query += " ORDER BY lr.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []LeaveRequest
	for rows.Next() {
		var r LeaveRequest
		if err := rows.Scan(&r.ID, &r.CompanyID, &r.EmployeeID, &r.EmployeeName,
			&r.LeaveTypeID, &r.LeaveTypeName, &r.StartDate, &r.EndDate, &r.TotalDays,
			&r.Reason, &r.AttachmentURL, &r.Status,
			&r.ReviewerID, &r.ReviewerComment, &r.ReviewedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Service) MyRequests(ctx context.Context, companyID, employeeID uuid.UUID) ([]LeaveRequest, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT lr.id, lr.company_id, lr.employee_id, '' as employee_name,
		        lr.leave_type_id, lt.name as leave_type_name,
		        lr.start_date::text, lr.end_date::text, lr.total_days,
		        lr.reason, lr.attachment_url, lr.status,
		        lr.reviewer_id, lr.reviewer_comment, lr.reviewed_at, lr.created_at
		 FROM leave_requests lr
		 JOIN leave_types lt ON lt.id = lr.leave_type_id
		 WHERE lr.company_id = $1 AND lr.employee_id = $2
		 ORDER BY lr.created_at DESC`, companyID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []LeaveRequest
	for rows.Next() {
		var r LeaveRequest
		if err := rows.Scan(&r.ID, &r.CompanyID, &r.EmployeeID, &r.EmployeeName,
			&r.LeaveTypeID, &r.LeaveTypeName, &r.StartDate, &r.EndDate, &r.TotalDays,
			&r.Reason, &r.AttachmentURL, &r.Status,
			&r.ReviewerID, &r.ReviewerComment, &r.ReviewedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Service) ReviewRequest(ctx context.Context, companyID, reqID, reviewerID uuid.UUID, status, comment string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var employeeID uuid.UUID
	var leaveTypeID uuid.UUID
	var totalDays int
	var startDate string
	err = tx.QueryRowContext(ctx,
		`UPDATE leave_requests SET status=$1, reviewer_id=$2, reviewer_comment=$3, reviewed_at=NOW()
		 WHERE id=$4 AND company_id=$5 AND status='pending'
		 RETURNING employee_id, leave_type_id, total_days, start_date::text`,
		status, reviewerID, comment, reqID, companyID,
	).Scan(&employeeID, &leaveTypeID, &totalDays, &startDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("request not found or already reviewed")
		}
		return err
	}

	if status == "approved" {
		res, err := tx.ExecContext(ctx,
			`UPDATE leave_balances SET used_days = used_days + $1
			 WHERE employee_id = $2 AND leave_type_id = $3 AND year = EXTRACT(YEAR FROM $4::date)
			 AND (total_days + carry_over_days - used_days) >= $1`,
			totalDays, employeeID, leaveTypeID, startDate)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return fmt.Errorf("insufficient leave balance remaining")
		}
	}

	return tx.Commit()
}

func (s *Service) CancelRequest(ctx context.Context, companyID, employeeID, reqID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var leaveTypeID uuid.UUID
	var totalDays int
	var startDate string
	var prevStatus string
	err = tx.QueryRowContext(ctx,
		`UPDATE leave_requests SET status='cancelled'
		 WHERE id=$1 AND company_id=$2 AND employee_id=$3 AND status IN ('pending','approved')
		 RETURNING leave_type_id, total_days, start_date::text, status`,
		reqID, companyID, employeeID,
	).Scan(&leaveTypeID, &totalDays, &startDate, &prevStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("request not found or cannot be cancelled")
		}
		return err
	}

	if prevStatus == "approved" {
		_, err = tx.ExecContext(ctx,
			`UPDATE leave_balances SET used_days = GREATEST(used_days - $1, 0)
			 WHERE employee_id = $2 AND leave_type_id = $3 AND year = EXTRACT(YEAR FROM $4::date)`,
			totalDays, employeeID, leaveTypeID, startDate)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
