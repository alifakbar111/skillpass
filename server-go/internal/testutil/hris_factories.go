package testutil

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func CreateDepartment(db *sql.DB, companyID uuid.UUID, name string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO departments (id, company_id, name, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())`,
		id, companyID, name,
	)
	return id, err
}

func CreatePosition(db *sql.DB, companyID uuid.UUID, name string, departmentID uuid.UUID) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO positions (id, company_id, name, department_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW())`,
		id, companyID, name, departmentID,
	)
	return id, err
}

func CreateBranch(db *sql.DB, companyID uuid.UUID, name string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO branches (id, company_id, name, branch_type, address, created_at, updated_at)
		 VALUES ($1, $2, $3, 'branch', '123 Test St', NOW(), NOW())`,
		id, companyID, name,
	)
	return id, err
}

func CreateEmployee(db *sql.DB, companyID uuid.UUID, firstName, lastName, email string) (uuid.UUID, error) {
	id := uuid.New()
	empNum := "EMP-" + id.String()[:8]
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO employees (id, company_id, employee_id_number, first_name, last_name, email,
		 employment_type, employment_status, join_date, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'permanent', 'active', $7, NOW(), NOW())`,
		id, companyID, empNum, firstName, lastName, email, time.Now().Format("2006-01-02"),
	)
	return id, err
}

func CreateAttendanceLog(db *sql.DB, companyID, employeeID uuid.UUID, date string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO attendance_logs (id, company_id, employee_id, date, clock_in, attendance_code, created_at)
		 VALUES ($1, $2, $3, $4, NOW(), 'P', NOW())`,
		id, companyID, employeeID, date,
	)
	return id, err
}

func CreateLeaveType(db *sql.DB, companyID uuid.UUID, name, code string, daysPerYear int) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO leave_types (id, company_id, name, code, default_days_per_year, is_paid, requires_attachment, is_active, created_at)
		 VALUES ($1, $2, $3, $4, $5, true, false, true, NOW())`,
		id, companyID, name, code, daysPerYear,
	)
	return id, err
}

func InitLeaveBalance(db *sql.DB, employeeID, leaveTypeID uuid.UUID, year, totalDays int) error {
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO leave_balances (id, employee_id, leave_type_id, year, total_days, used_days, carry_over_days, created_at)
		 VALUES ($1, $2, $3, $4, $5, 0, 0, NOW())
		 ON CONFLICT (employee_id, leave_type_id, year) DO NOTHING`,
		uuid.New(), employeeID, leaveTypeID, year, totalDays,
	)
	return err
}

func CreateHoliday(db *sql.DB, companyID uuid.UUID, name, date string, isRecurring bool) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO holidays (id, company_id, name, date, is_recurring, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())`,
		id, companyID, name, date, isRecurring,
	)
	return id, err
}
