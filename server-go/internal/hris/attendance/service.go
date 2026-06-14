package attendance

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

type AttendanceLog struct {
	ID              uuid.UUID  `json:"id"`
	CompanyID       uuid.UUID  `json:"companyId"`
	EmployeeID      uuid.UUID  `json:"employeeId"`
	EmployeeName    string     `json:"employeeName,omitempty"`
	Date            string     `json:"date"`
	ClockIn         *time.Time `json:"clockIn"`
	ClockOut        *time.Time `json:"clockOut"`
	ClockInLat      *float64   `json:"clockInLat"`
	ClockInLng      *float64   `json:"clockInLng"`
	ClockOutLat     *float64   `json:"clockOutLat"`
	ClockOutLng     *float64   `json:"clockOutLng"`
	BranchID        *uuid.UUID `json:"branchId"`
	IsInGeofence    *bool      `json:"isInGeofence"`
	IsLate          bool       `json:"isLate"`
	LateMinutes     int        `json:"lateMinutes"`
	IsEarlyOut      bool       `json:"isEarlyOut"`
	OvertimeMinutes int        `json:"overtimeMinutes"`
	AttendanceCode  string     `json:"attendanceCode"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type ClockInRequest struct {
	Lat      float64    `json:"lat"`
	Lng      float64    `json:"lng"`
	BranchID *uuid.UUID `json:"branchId"`
}

type ClockOutRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type AttendanceException struct {
	ID              uuid.UUID  `json:"id"`
	CompanyID       uuid.UUID  `json:"companyId"`
	EmployeeID      uuid.UUID  `json:"employeeId"`
	EmployeeName    string     `json:"employeeName,omitempty"`
	Date            string     `json:"date"`
	ExceptionType   string     `json:"exceptionType"`
	Reason          string     `json:"reason"`
	AttachmentURL   *string    `json:"attachmentUrl"`
	Status          string     `json:"status"`
	ReviewerID      *uuid.UUID `json:"reviewerId"`
	ReviewerComment *string    `json:"reviewerComment"`
	ReviewedAt      *time.Time `json:"reviewedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type DashboardStats struct {
	TotalEmployees int `json:"totalEmployees"`
	Present        int `json:"present"`
	Late           int `json:"late"`
	Absent         int `json:"absent"`
	OnLeave        int `json:"onLeave"`
}

func (s *Service) ClockIn(ctx context.Context, companyID, employeeID uuid.UUID, req ClockInRequest) (*AttendanceLog, error) {
	today := time.Now().Format("2006-01-02")

	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM attendance_logs WHERE employee_id=$1 AND date=$2)`,
		employeeID, today).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("already clocked in today")
	}

	var inGeofence *bool
	if req.BranchID != nil {
		var branchLat, branchLng, radius float64
		err := s.db.QueryRowContext(ctx,
			`SELECT COALESCE(latitude,0), COALESCE(longitude,0), COALESCE(geofence_radius_meters,100)
			 FROM branches WHERE id=$1 AND company_id=$2`, req.BranchID, companyID,
		).Scan(&branchLat, &branchLng, &radius)
		if err == nil && branchLat != 0 {
			result := IsWithinGeofence(req.Lat, req.Lng, branchLat, branchLng, radius)
			inGeofence = &result
		}
	}

	isLate := false
	lateMinutes := 0
	err = s.db.QueryRowContext(ctx,
		`SELECT st.start_time, st.late_tolerance_minutes
		 FROM employee_shifts es
		 JOIN shift_templates st ON st.id = es.shift_id
		 WHERE es.employee_id = $1
		   AND es.effective_date <= $2
		   AND (es.end_date IS NULL OR es.end_date >= $2)
		 ORDER BY es.effective_date DESC LIMIT 1`, employeeID, today,
	).Scan(new(string), new(int))
	if err == nil {
		var startTime string
		var tolerance int
		_ = s.db.QueryRowContext(ctx,
			`SELECT st.start_time::text, st.late_tolerance_minutes
			 FROM employee_shifts es
			 JOIN shift_templates st ON st.id = es.shift_id
			 WHERE es.employee_id = $1
			   AND es.effective_date <= $2
			   AND (es.end_date IS NULL OR es.end_date >= $2)
			 ORDER BY es.effective_date DESC LIMIT 1`, employeeID, today,
		).Scan(&startTime, &tolerance)
		shiftStart, parseErr := time.Parse("15:04:05", startTime)
		if parseErr == nil {
			now := time.Now()
			shiftStartToday := time.Date(now.Year(), now.Month(), now.Day(),
				shiftStart.Hour(), shiftStart.Minute(), 0, 0, now.Location())
			deadline := shiftStartToday.Add(time.Duration(tolerance) * time.Minute)
			if now.After(deadline) {
				isLate = true
				lateMinutes = int(now.Sub(shiftStartToday).Minutes())
			}
		}
	}

	var log AttendanceLog
	clockIn := time.Now()
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO attendance_logs
		  (company_id, employee_id, date, clock_in, clock_in_lat, clock_in_lng,
		   branch_id, is_in_geofence, is_late, late_minutes, attendance_code)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'P')
		 RETURNING id, company_id, employee_id, date::text, clock_in,
		           clock_in_lat, clock_in_lng, branch_id, is_in_geofence,
		           is_late, late_minutes, attendance_code, created_at`,
		companyID, employeeID, today, clockIn, req.Lat, req.Lng,
		req.BranchID, inGeofence, isLate, lateMinutes,
	).Scan(&log.ID, &log.CompanyID, &log.EmployeeID, &log.Date, &log.ClockIn,
		&log.ClockInLat, &log.ClockInLng, &log.BranchID, &log.IsInGeofence,
		&log.IsLate, &log.LateMinutes, &log.AttendanceCode, &log.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (s *Service) ClockOut(ctx context.Context, companyID, employeeID uuid.UUID, req ClockOutRequest) (*AttendanceLog, error) {
	today := time.Now().Format("2006-01-02")
	clockOut := time.Now()

	var log AttendanceLog
	err := s.db.QueryRowContext(ctx,
		`UPDATE attendance_logs SET clock_out=$1, clock_out_lat=$2, clock_out_lng=$3
		 WHERE employee_id=$4 AND company_id=$5 AND date=$6 AND clock_out IS NULL
		 RETURNING id, company_id, employee_id, date::text, clock_in, clock_out,
		           clock_in_lat, clock_in_lng, clock_out_lat, clock_out_lng,
		           branch_id, is_in_geofence, is_late, late_minutes,
		           is_early_out, overtime_minutes, attendance_code, created_at`,
		clockOut, req.Lat, req.Lng, employeeID, companyID, today,
	).Scan(&log.ID, &log.CompanyID, &log.EmployeeID, &log.Date,
		&log.ClockIn, &log.ClockOut, &log.ClockInLat, &log.ClockInLng,
		&log.ClockOutLat, &log.ClockOutLng, &log.BranchID, &log.IsInGeofence,
		&log.IsLate, &log.LateMinutes, &log.IsEarlyOut, &log.OvertimeMinutes,
		&log.AttendanceCode, &log.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active clock-in found for today")
		}
		return nil, err
	}
	return &log, nil
}

func (s *Service) GetDashboard(ctx context.Context, companyID uuid.UUID, date string) (*DashboardStats, []AttendanceLog, error) {
	var stats DashboardStats
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM employees WHERE company_id=$1 AND employment_status='active'`,
		companyID).Scan(&stats.TotalEmployees)
	if err != nil {
		return nil, nil, err
	}

	err = s.db.QueryRowContext(ctx,
		`SELECT
		   COALESCE(SUM(CASE WHEN attendance_code='P' AND NOT is_late THEN 1 ELSE 0 END),0),
		   COALESCE(SUM(CASE WHEN is_late THEN 1 ELSE 0 END),0)
		 FROM attendance_logs WHERE company_id=$1 AND date=$2`,
		companyID, date).Scan(&stats.Present, &stats.Late)
	if err != nil {
		return nil, nil, err
	}

	stats.Absent = stats.TotalEmployees - stats.Present - stats.Late

	rows, err := s.db.QueryContext(ctx,
		`SELECT a.id, a.company_id, a.employee_id,
		        COALESCE(e.first_name||' '||e.last_name, '') as employee_name,
		        a.date::text, a.clock_in, a.clock_out,
		        a.clock_in_lat, a.clock_in_lng, a.clock_out_lat, a.clock_out_lng,
		        a.branch_id, a.is_in_geofence,
		        a.is_late, a.late_minutes, a.is_early_out, a.overtime_minutes,
		        a.attendance_code, a.created_at
		 FROM attendance_logs a
		 JOIN employees e ON e.id = a.employee_id
		 WHERE a.company_id=$1 AND a.date=$2
		 ORDER BY a.clock_in DESC NULLS LAST`, companyID, date)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var logs []AttendanceLog
	for rows.Next() {
		var l AttendanceLog
		if err := rows.Scan(&l.ID, &l.CompanyID, &l.EmployeeID, &l.EmployeeName,
			&l.Date, &l.ClockIn, &l.ClockOut,
			&l.ClockInLat, &l.ClockInLng, &l.ClockOutLat, &l.ClockOutLng,
			&l.BranchID, &l.IsInGeofence,
			&l.IsLate, &l.LateMinutes, &l.IsEarlyOut, &l.OvertimeMinutes,
			&l.AttendanceCode, &l.CreatedAt); err != nil {
			return nil, nil, err
		}
		logs = append(logs, l)
	}
	return &stats, logs, rows.Err()
}

func (s *Service) GetMyAttendance(ctx context.Context, companyID, employeeID uuid.UUID, month string) ([]AttendanceLog, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, employee_id, date::text, clock_in, clock_out,
		        clock_in_lat, clock_in_lng, clock_out_lat, clock_out_lng,
		        branch_id, is_in_geofence,
		        is_late, late_minutes, is_early_out, overtime_minutes,
		        attendance_code, created_at
		 FROM attendance_logs
		 WHERE company_id=$1 AND employee_id=$2 AND to_char(date,'YYYY-MM')=$3
		 ORDER BY date DESC`, companyID, employeeID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AttendanceLog
	for rows.Next() {
		var l AttendanceLog
		if err := rows.Scan(&l.ID, &l.CompanyID, &l.EmployeeID, &l.Date,
			&l.ClockIn, &l.ClockOut, &l.ClockInLat, &l.ClockInLng,
			&l.ClockOutLat, &l.ClockOutLng, &l.BranchID, &l.IsInGeofence,
			&l.IsLate, &l.LateMinutes, &l.IsEarlyOut, &l.OvertimeMinutes,
			&l.AttendanceCode, &l.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

func (s *Service) CreateException(ctx context.Context, companyID, employeeID uuid.UUID, ex *AttendanceException) error {
	return s.db.QueryRowContext(ctx,
		`INSERT INTO attendance_exceptions (company_id, employee_id, date, exception_type, reason, attachment_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, status, created_at`,
		companyID, employeeID, ex.Date, ex.ExceptionType, ex.Reason, ex.AttachmentURL,
	).Scan(&ex.ID, &ex.Status, &ex.CreatedAt)
}

func (s *Service) ListExceptions(ctx context.Context, companyID uuid.UUID, status string) ([]AttendanceException, error) {
	query := `SELECT ae.id, ae.company_id, ae.employee_id,
	                 COALESCE(e.first_name||' '||e.last_name,'') as employee_name,
	                 ae.date::text, ae.exception_type, ae.reason, ae.attachment_url,
	                 ae.status, ae.reviewer_id, ae.reviewer_comment, ae.reviewed_at, ae.created_at
	          FROM attendance_exceptions ae
	          JOIN employees e ON e.id = ae.employee_id
	          WHERE ae.company_id=$1`
	args := []any{companyID}
	if status != "" {
		query += " AND ae.status=$2"
		args = append(args, status)
	}
	query += " ORDER BY ae.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AttendanceException
	for rows.Next() {
		var ex AttendanceException
		if err := rows.Scan(&ex.ID, &ex.CompanyID, &ex.EmployeeID, &ex.EmployeeName,
			&ex.Date, &ex.ExceptionType, &ex.Reason, &ex.AttachmentURL,
			&ex.Status, &ex.ReviewerID, &ex.ReviewerComment, &ex.ReviewedAt,
			&ex.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, ex)
	}
	return list, rows.Err()
}

func (s *Service) ReviewException(ctx context.Context, companyID, exID, reviewerID uuid.UUID, status, comment string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE attendance_exceptions SET status=$1, reviewer_id=$2, reviewer_comment=$3, reviewed_at=NOW()
		 WHERE id=$4 AND company_id=$5 AND status='pending'`,
		status, reviewerID, comment, exID, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("exception not found or already reviewed")
	}
	return nil
}
