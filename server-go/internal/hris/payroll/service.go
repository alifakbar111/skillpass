package payroll

import (
	"context"
	"database/sql"
	"encoding/json"
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

type SalaryComponent struct {
	ID            uuid.UUID `json:"id"`
	CompanyID     uuid.UUID `json:"companyId"`
	Name          string    `json:"name"`
	Code          string    `json:"code"`
	Type          string    `json:"type"`
	IsTaxable     bool      `json:"isTaxable"`
	IsFixed       bool      `json:"isFixed"`
	DefaultAmount float64   `json:"defaultAmount"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
}

type EmployeeSalary struct {
	ID            uuid.UUID `json:"id"`
	EmployeeID    uuid.UUID `json:"employeeId"`
	ComponentID   uuid.UUID `json:"componentId"`
	ComponentName string    `json:"componentName,omitempty"`
	ComponentCode string    `json:"componentCode,omitempty"`
	ComponentType string    `json:"componentType,omitempty"`
	Amount        float64   `json:"amount"`
	EffectiveDate string    `json:"effectiveDate"`
	CreatedAt     time.Time `json:"createdAt"`
}

type PayrollRun struct {
	ID              uuid.UUID  `json:"id"`
	CompanyID       uuid.UUID  `json:"companyId"`
	PeriodStart     string     `json:"periodStart"`
	PeriodEnd       string     `json:"periodEnd"`
	Status          string     `json:"status"`
	TotalGross      float64    `json:"totalGross"`
	TotalDeductions float64    `json:"totalDeductions"`
	TotalNet        float64    `json:"totalNet"`
	EmployeeCount   int        `json:"employeeCount"`
	Notes           *string    `json:"notes"`
	RunBy           *uuid.UUID `json:"runBy"`
	RunByName       string     `json:"runByName,omitempty"`
	ApprovedBy      *uuid.UUID `json:"approvedBy"`
	ApprovedAt      *time.Time `json:"approvedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type PayslipLine struct {
	ComponentName string  `json:"componentName"`
	ComponentCode string  `json:"componentCode"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
}

type Payslip struct {
	ID              uuid.UUID     `json:"id"`
	PayrollRunID    uuid.UUID     `json:"payrollRunId"`
	EmployeeID      uuid.UUID     `json:"employeeId"`
	EmployeeName    string        `json:"employeeName,omitempty"`
	EmployeeCode    string        `json:"employeeCode,omitempty"`
	GrossPay        float64       `json:"grossPay"`
	TotalDeductions float64       `json:"totalDeductions"`
	NetPay          float64       `json:"netPay"`
	Breakdown       []PayslipLine `json:"breakdown"`
	CreatedAt       time.Time     `json:"createdAt"`
	PeriodStart     string        `json:"periodStart,omitempty"`
	PeriodEnd       string        `json:"periodEnd,omitempty"`
}

// ── Salary Components ──

func (s *Service) ListComponents(ctx context.Context, companyID uuid.UUID) ([]SalaryComponent, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, code, type, is_taxable, is_fixed, default_amount, is_active, created_at
		 FROM salary_components WHERE company_id = $1 ORDER BY type, name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SalaryComponent
	for rows.Next() {
		var c SalaryComponent
		if err := rows.Scan(&c.ID, &c.CompanyID, &c.Name, &c.Code, &c.Type, &c.IsTaxable,
			&c.IsFixed, &c.DefaultAmount, &c.IsActive, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *Service) CreateComponent(ctx context.Context, companyID uuid.UUID, c *SalaryComponent) error {
	return s.db.QueryRowContext(ctx,
		`INSERT INTO salary_components (company_id, name, code, type, is_taxable, is_fixed, default_amount)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, is_active, created_at`,
		companyID, c.Name, c.Code, c.Type, c.IsTaxable, c.IsFixed, c.DefaultAmount,
	).Scan(&c.ID, &c.IsActive, &c.CreatedAt)
}

func (s *Service) UpdateComponent(ctx context.Context, companyID, id uuid.UUID, c *SalaryComponent) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE salary_components SET name=$1, code=$2, type=$3, is_taxable=$4, is_fixed=$5, default_amount=$6, is_active=$7
		 WHERE id=$8 AND company_id=$9`,
		c.Name, c.Code, c.Type, c.IsTaxable, c.IsFixed, c.DefaultAmount, c.IsActive, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Service) DeleteComponent(ctx context.Context, companyID, id uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM salary_components WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ── Employee Salary ──

func (s *Service) GetEmployeeSalary(ctx context.Context, companyID, employeeID uuid.UUID) ([]EmployeeSalary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT es.id, es.employee_id, es.component_id, sc.name, sc.code, sc.type,
		        es.amount, es.effective_date::text, es.created_at
		 FROM employee_salary es
		 JOIN salary_components sc ON sc.id = es.component_id
		 JOIN employees e ON e.id = es.employee_id
		 WHERE e.company_id = $1 AND es.employee_id = $2
		 ORDER BY sc.type, sc.name`, companyID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []EmployeeSalary
	for rows.Next() {
		var es EmployeeSalary
		if err := rows.Scan(&es.ID, &es.EmployeeID, &es.ComponentID, &es.ComponentName,
			&es.ComponentCode, &es.ComponentType, &es.Amount, &es.EffectiveDate, &es.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, es)
	}
	return list, rows.Err()
}

func (s *Service) SetEmployeeSalary(ctx context.Context, companyID, employeeID uuid.UUID, items []EmployeeSalary) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM employees WHERE id=$1 AND company_id=$2)`, employeeID, companyID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("employee not found")
	}

	for _, item := range items {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO employee_salary (employee_id, component_id, amount, effective_date)
			 VALUES ($1, $2, $3, COALESCE(NULLIF($4, '')::date, CURRENT_DATE))
			 ON CONFLICT (employee_id, component_id) DO UPDATE SET amount = $3, effective_date = COALESCE(NULLIF($4, '')::date, CURRENT_DATE)`,
			employeeID, item.ComponentID, item.Amount, item.EffectiveDate)
		if err != nil {
			return fmt.Errorf("failed to set component %s: %w", item.ComponentID, err)
		}
	}

	return tx.Commit()
}

// ── Payroll Runs ──

func (s *Service) ListRuns(ctx context.Context, companyID uuid.UUID) ([]PayrollRun, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT pr.id, pr.company_id, pr.period_start::text, pr.period_end::text,
		        pr.status, pr.total_gross, pr.total_deductions, pr.total_net,
		        pr.employee_count, pr.notes, pr.run_by,
		        COALESCE(e.first_name||' '||e.last_name, '') as run_by_name,
		        pr.approved_by, pr.approved_at, pr.created_at
		 FROM payroll_runs pr
		 LEFT JOIN employees e ON e.id = pr.run_by
		 WHERE pr.company_id = $1
		 ORDER BY pr.created_at DESC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PayrollRun
	for rows.Next() {
		var r PayrollRun
		if err := rows.Scan(&r.ID, &r.CompanyID, &r.PeriodStart, &r.PeriodEnd,
			&r.Status, &r.TotalGross, &r.TotalDeductions, &r.TotalNet,
			&r.EmployeeCount, &r.Notes, &r.RunBy, &r.RunByName,
			&r.ApprovedBy, &r.ApprovedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Service) CreateRun(ctx context.Context, companyID, runBy uuid.UUID, periodStart, periodEnd string, notes *string) (*PayrollRun, error) {
	var r PayrollRun
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO payroll_runs (company_id, period_start, period_end, notes, run_by)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, company_id, period_start::text, period_end::text, status,
		           total_gross, total_deductions, total_net, employee_count, notes,
		           run_by, approved_by, approved_at, created_at`,
		companyID, periodStart, periodEnd, notes, runBy,
	).Scan(&r.ID, &r.CompanyID, &r.PeriodStart, &r.PeriodEnd, &r.Status,
		&r.TotalGross, &r.TotalDeductions, &r.TotalNet, &r.EmployeeCount, &r.Notes,
		&r.RunBy, &r.ApprovedBy, &r.ApprovedAt, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) CalculateRun(ctx context.Context, companyID, runID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM payroll_runs WHERE id=$1 AND company_id=$2`, runID, companyID).Scan(&status)
	if err != nil {
		return fmt.Errorf("payroll run not found")
	}
	if status != "draft" {
		return fmt.Errorf("can only calculate runs in draft status")
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM payslips WHERE payroll_run_id = $1`, runID)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx,
		`SELECT e.id, e.employee_code, e.first_name, e.last_name
		 FROM employees e
		 WHERE e.company_id = $1 AND e.status = 'active'
		 ORDER BY e.first_name, e.last_name`, companyID)
	if err != nil {
		return err
	}

	type empInfo struct {
		id        uuid.UUID
		code      string
		firstName string
		lastName  string
	}
	var employees []empInfo
	for rows.Next() {
		var e empInfo
		if err := rows.Scan(&e.id, &e.code, &e.firstName, &e.lastName); err != nil {
			rows.Close()
			return err
		}
		employees = append(employees, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	var totalGross, totalDeductions, totalNet float64
	employeeCount := 0

	for _, emp := range employees {
		salaryRows, err := tx.QueryContext(ctx,
			`SELECT sc.name, sc.code, sc.type, es.amount
			 FROM employee_salary es
			 JOIN salary_components sc ON sc.id = es.component_id
			 WHERE es.employee_id = $1 AND sc.is_active = true
			 ORDER BY sc.type, sc.name`, emp.id)
		if err != nil {
			return err
		}

		var lines []PayslipLine
		var gross, deductions float64
		for salaryRows.Next() {
			var line PayslipLine
			if err := salaryRows.Scan(&line.ComponentName, &line.ComponentCode, &line.Type, &line.Amount); err != nil {
				salaryRows.Close()
				return err
			}
			lines = append(lines, line)
			if line.Type == "earning" {
				gross += line.Amount
			} else {
				deductions += line.Amount
			}
		}
		salaryRows.Close()
		if err := salaryRows.Err(); err != nil {
			return err
		}

		if len(lines) == 0 {
			continue
		}

		net := gross - deductions
		breakdownJSON, err := json.Marshal(lines)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO payslips (payroll_run_id, employee_id, gross_pay, total_deductions, net_pay, breakdown)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			runID, emp.id, gross, deductions, net, breakdownJSON)
		if err != nil {
			return err
		}

		totalGross += gross
		totalDeductions += deductions
		totalNet += net
		employeeCount++
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE payroll_runs SET status='calculated', total_gross=$1, total_deductions=$2, total_net=$3, employee_count=$4
		 WHERE id=$5`,
		totalGross, totalDeductions, totalNet, employeeCount, runID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) ApproveRun(ctx context.Context, companyID, runID, approverID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE payroll_runs SET status='approved', approved_by=$1, approved_at=NOW()
		 WHERE id=$2 AND company_id=$3 AND status='calculated'`,
		approverID, runID, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("run not found or not in calculated status")
	}
	return nil
}

func (s *Service) MarkPaid(ctx context.Context, companyID, runID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE payroll_runs SET status='paid'
		 WHERE id=$1 AND company_id=$2 AND status='approved'`,
		runID, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("run not found or not in approved status")
	}
	return nil
}

// ── Payslips ──

func (s *Service) ListPayslips(ctx context.Context, companyID, runID uuid.UUID) ([]Payslip, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.payroll_run_id, p.employee_id,
		        COALESCE(e.first_name||' '||e.last_name, '') as employee_name,
		        e.employee_code,
		        p.gross_pay, p.total_deductions, p.net_pay, p.breakdown, p.created_at,
		        pr.period_start::text, pr.period_end::text
		 FROM payslips p
		 JOIN employees e ON e.id = p.employee_id
		 JOIN payroll_runs pr ON pr.id = p.payroll_run_id
		 WHERE pr.company_id = $1 AND p.payroll_run_id = $2
		 ORDER BY e.first_name, e.last_name`, companyID, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Payslip
	for rows.Next() {
		var p Payslip
		var breakdownJSON []byte
		if err := rows.Scan(&p.ID, &p.PayrollRunID, &p.EmployeeID, &p.EmployeeName, &p.EmployeeCode,
			&p.GrossPay, &p.TotalDeductions, &p.NetPay, &breakdownJSON, &p.CreatedAt,
			&p.PeriodStart, &p.PeriodEnd); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(breakdownJSON, &p.Breakdown); err != nil {
			p.Breakdown = []PayslipLine{}
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Service) GetMyPayslips(ctx context.Context, companyID, employeeID uuid.UUID) ([]Payslip, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.payroll_run_id, p.employee_id, '' as employee_name, '' as employee_code,
		        p.gross_pay, p.total_deductions, p.net_pay, p.breakdown, p.created_at,
		        pr.period_start::text, pr.period_end::text
		 FROM payslips p
		 JOIN payroll_runs pr ON pr.id = p.payroll_run_id
		 WHERE pr.company_id = $1 AND p.employee_id = $2 AND pr.status IN ('approved', 'paid')
		 ORDER BY pr.period_start DESC`, companyID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Payslip
	for rows.Next() {
		var p Payslip
		var breakdownJSON []byte
		if err := rows.Scan(&p.ID, &p.PayrollRunID, &p.EmployeeID, &p.EmployeeName, &p.EmployeeCode,
			&p.GrossPay, &p.TotalDeductions, &p.NetPay, &breakdownJSON, &p.CreatedAt,
			&p.PeriodStart, &p.PeriodEnd); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(breakdownJSON, &p.Breakdown); err != nil {
			p.Breakdown = []PayslipLine{}
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Service) GetPayslip(ctx context.Context, companyID, payslipID uuid.UUID) (*Payslip, error) {
	var p Payslip
	var breakdownJSON []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT p.id, p.payroll_run_id, p.employee_id,
		        COALESCE(e.first_name||' '||e.last_name, '') as employee_name,
		        e.employee_code,
		        p.gross_pay, p.total_deductions, p.net_pay, p.breakdown, p.created_at,
		        pr.period_start::text, pr.period_end::text
		 FROM payslips p
		 JOIN employees e ON e.id = p.employee_id
		 JOIN payroll_runs pr ON pr.id = p.payroll_run_id
		 WHERE pr.company_id = $1 AND p.id = $2`,
		companyID, payslipID).Scan(&p.ID, &p.PayrollRunID, &p.EmployeeID, &p.EmployeeName, &p.EmployeeCode,
		&p.GrossPay, &p.TotalDeductions, &p.NetPay, &breakdownJSON, &p.CreatedAt,
		&p.PeriodStart, &p.PeriodEnd)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(breakdownJSON, &p.Breakdown); err != nil {
		p.Breakdown = []PayslipLine{}
	}
	return &p, nil
}
