package report

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type AttendanceRow struct {
	EmployeeName string `json:"employeeName"`
	EmployeeCode string `json:"employeeCode"`
	Date         string `json:"date"`
	ClockIn      string `json:"clockIn"`
	ClockOut     string `json:"clockOut"`
	WorkHours    string `json:"workHours"`
	Status       string `json:"status"`
	ShiftName    string `json:"shiftName"`
}

type DeptBreakdown struct {
	Department string `json:"department"`
	Headcount  int    `json:"headcount"`
	NewHires   int    `json:"newHires"`
	Exits      int    `json:"exits"`
}

type AnalyticsSnapshot struct {
	ID                  uuid.UUID       `json:"id"`
	SnapshotMonth       string          `json:"snapshotMonth"`
	TotalHeadcount      int             `json:"totalHeadcount"`
	NewHires            int             `json:"newHires"`
	Terminations        int             `json:"terminations"`
	TurnoverRate        float64         `json:"turnoverRate"`
	AvgTenureMonths     float64         `json:"avgTenureMonths"`
	DepartmentBreakdown []DeptBreakdown `json:"departmentBreakdown"`
	CreatedAt           time.Time       `json:"createdAt"`
}

type HeadcountStats struct {
	TotalActive    int             `json:"totalActive"`
	ByDepartment   []DeptCount     `json:"byDepartment"`
	ByBranch       []BranchCount   `json:"byBranch"`
	ByStatus       []StatusCount   `json:"byStatus"`
	AvgTenure      float64         `json:"avgTenureMonths"`
	GenderBreakdown []GenderCount  `json:"genderBreakdown"`
}

type DeptCount struct {
	Department string `json:"department"`
	Count      int    `json:"count"`
}

type BranchCount struct {
	Branch string `json:"branch"`
	Count  int    `json:"count"`
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type GenderCount struct {
	Gender string `json:"gender"`
	Count  int    `json:"count"`
}

func (s *Service) ExportAttendance(ctx context.Context, companyID uuid.UUID, from, to string) ([]AttendanceRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(e.first_name||' '||e.last_name, '') as employee_name,
			e.employee_id_number,
			al.date::text,
			COALESCE(TO_CHAR(al.clock_in, 'HH24:MI'), ''),
			COALESCE(TO_CHAR(al.clock_out, 'HH24:MI'), ''),
			COALESCE(
				ROUND(EXTRACT(EPOCH FROM (al.clock_out - al.clock_in))/3600, 2)::text,
				''
			),
			al.attendance_code,
			''
		FROM attendance_logs al
		JOIN employees e ON e.id = al.employee_id AND e.company_id = $1
		WHERE al.company_id = $1
			AND al.date >= $2::date
			AND al.date <= $3::date
		ORDER BY al.date, e.first_name, e.last_name
	`, companyID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AttendanceRow
	for rows.Next() {
		var r AttendanceRow
		if err := rows.Scan(&r.EmployeeName, &r.EmployeeCode, &r.Date, &r.ClockIn, &r.ClockOut, &r.WorkHours, &r.Status, &r.ShiftName); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *Service) ToCSV(rows []AttendanceRow) string {
	var b strings.Builder
	b.WriteString("Employee Name,Employee Code,Date,Clock In,Clock Out,Work Hours,Status,Shift\n")
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			csvEscape(r.EmployeeName), r.EmployeeCode, r.Date, r.ClockIn, r.ClockOut, r.WorkHours, r.Status, csvEscape(r.ShiftName)))
	}
	return b.String()
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func (s *Service) GetHeadcountStats(ctx context.Context, companyID uuid.UUID) (*HeadcountStats, error) {
	stats := &HeadcountStats{}

	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM employees WHERE company_id = $1 AND employment_status = 'active'`,
		companyID).Scan(&stats.TotalActive)
	if err != nil {
		return nil, err
	}

	deptRows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(d.name, 'Unassigned'), COUNT(*)
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE e.company_id = $1 AND e.employment_status = 'active'
		GROUP BY d.name ORDER BY COUNT(*) DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer deptRows.Close()
	for deptRows.Next() {
		var dc DeptCount
		if err := deptRows.Scan(&dc.Department, &dc.Count); err != nil {
			return nil, err
		}
		stats.ByDepartment = append(stats.ByDepartment, dc)
	}
	if err := deptRows.Err(); err != nil {
		return nil, err
	}

	branchRows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(b.name, 'Unassigned'), COUNT(*)
		FROM employees e
		LEFT JOIN branches b ON b.id = e.branch_id
		WHERE e.company_id = $1 AND e.employment_status = 'active'
		GROUP BY b.name ORDER BY COUNT(*) DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer branchRows.Close()
	for branchRows.Next() {
		var bc BranchCount
		if err := branchRows.Scan(&bc.Branch, &bc.Count); err != nil {
			return nil, err
		}
		stats.ByBranch = append(stats.ByBranch, bc)
	}
	if err := branchRows.Err(); err != nil {
		return nil, err
	}

	statusRows, err := s.db.QueryContext(ctx, `
		SELECT employment_status, COUNT(*)
		FROM employees WHERE company_id = $1
		GROUP BY employment_status ORDER BY COUNT(*) DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var sc StatusCount
		if err := statusRows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		stats.ByStatus = append(stats.ByStatus, sc)
	}
	if err := statusRows.Err(); err != nil {
		return nil, err
	}

	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (now() - join_date))/2592000), 0)
		FROM employees WHERE company_id = $1 AND employment_status = 'active' AND join_date IS NOT NULL
	`, companyID).Scan(&stats.AvgTenure)

	genderRows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(NULLIF(gender, ''), 'Not specified'), COUNT(*)
		FROM employees WHERE company_id = $1 AND employment_status = 'active'
		GROUP BY gender ORDER BY COUNT(*) DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer genderRows.Close()
	for genderRows.Next() {
		var gc GenderCount
		if err := genderRows.Scan(&gc.Gender, &gc.Count); err != nil {
			return nil, err
		}
		stats.GenderBreakdown = append(stats.GenderBreakdown, gc)
	}

	return stats, genderRows.Err()
}

func (s *Service) GenerateSnapshot(ctx context.Context, companyID uuid.UUID, month string) (*AnalyticsSnapshot, error) {
	monthStart := month + "-01"
	t, _ := time.Parse("2006-01-02", monthStart)
	monthEnd := t.AddDate(0, 1, -1).Format("2006-01-02")

	var totalHeadcount, newHires, terminations int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM employees WHERE company_id = $1 AND employment_status = 'active'
		 AND (join_date IS NULL OR join_date <= ($2::date + interval '1 month' - interval '1 day'))`,
		companyID, monthStart).Scan(&totalHeadcount); err != nil {
		return nil, err
	}

	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM employees WHERE company_id = $1
		 AND join_date >= $2::date AND join_date < $2::date + interval '1 month'`,
		companyID, monthStart).Scan(&newHires); err != nil {
		return nil, err
	}

	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM employees WHERE company_id = $1 AND employment_status = 'terminated'
		 AND updated_at >= $2::date AND updated_at < $2::date + interval '1 month'`,
		companyID, monthStart).Scan(&terminations); err != nil {
		return nil, err
	}

	var turnoverRate float64
	if totalHeadcount > 0 {
		turnoverRate = float64(terminations) / float64(totalHeadcount) * 100
	}

	var avgTenure float64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(AVG(EXTRACT(EPOCH FROM ($2::date - join_date))/2592000), 0)
		 FROM employees WHERE company_id = $1 AND employment_status = 'active' AND join_date IS NOT NULL`,
		companyID, monthEnd).Scan(&avgTenure); err != nil {
		return nil, err
	}

	deptRows, err := s.db.QueryContext(ctx, `
		SELECT COALESCE(d.name, 'Unassigned'),
			COUNT(*) FILTER (WHERE e.employment_status = 'active'),
			COUNT(*) FILTER (WHERE e.join_date >= $2::date AND e.join_date < $2::date + interval '1 month'),
			COUNT(*) FILTER (WHERE e.employment_status = 'terminated' AND e.updated_at >= $2::date AND e.updated_at < $2::date + interval '1 month')
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE e.company_id = $1
		GROUP BY d.name
	`, companyID, monthStart)
	if err != nil {
		return nil, err
	}
	defer deptRows.Close()

	var breakdown []DeptBreakdown
	for deptRows.Next() {
		var db DeptBreakdown
		if err := deptRows.Scan(&db.Department, &db.Headcount, &db.NewHires, &db.Exits); err != nil {
			return nil, err
		}
		breakdown = append(breakdown, db)
	}
	if err := deptRows.Err(); err != nil {
		return nil, err
	}

	breakdownJSON, err := json.Marshal(breakdown)
	if err != nil {
		return nil, err
	}

	snap := &AnalyticsSnapshot{}
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO hr_analytics_snapshots
			(company_id, snapshot_month, total_headcount, new_hires, terminations, turnover_rate, avg_tenure_months, department_breakdown)
		VALUES ($1, $2::date, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (company_id, snapshot_month) DO UPDATE SET
			total_headcount = EXCLUDED.total_headcount,
			new_hires = EXCLUDED.new_hires,
			terminations = EXCLUDED.terminations,
			turnover_rate = EXCLUDED.turnover_rate,
			avg_tenure_months = EXCLUDED.avg_tenure_months,
			department_breakdown = EXCLUDED.department_breakdown,
			created_at = now()
		RETURNING id, snapshot_month::text, total_headcount, new_hires, terminations, turnover_rate, avg_tenure_months, created_at
	`, companyID, monthStart, totalHeadcount, newHires, terminations, turnoverRate, avgTenure, breakdownJSON).Scan(
		&snap.ID, &snap.SnapshotMonth, &snap.TotalHeadcount, &snap.NewHires, &snap.Terminations,
		&snap.TurnoverRate, &snap.AvgTenureMonths, &snap.CreatedAt)
	if err != nil {
		return nil, err
	}
	snap.DepartmentBreakdown = breakdown
	return snap, nil
}

func (s *Service) ListSnapshots(ctx context.Context, companyID uuid.UUID) ([]AnalyticsSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, snapshot_month::text, total_headcount, new_hires, terminations,
			turnover_rate, avg_tenure_months, department_breakdown, created_at
		FROM hr_analytics_snapshots
		WHERE company_id = $1
		ORDER BY snapshot_month DESC
		LIMIT 24
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AnalyticsSnapshot
	for rows.Next() {
		var s AnalyticsSnapshot
		var breakdownRaw []byte
		if err := rows.Scan(&s.ID, &s.SnapshotMonth, &s.TotalHeadcount, &s.NewHires, &s.Terminations,
			&s.TurnoverRate, &s.AvgTenureMonths, &breakdownRaw, &s.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(breakdownRaw, &s.DepartmentBreakdown); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
