package employee

import (
	"context"
	"database/sql"
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

type Employee struct {
	ID                      uuid.UUID  `json:"id"`
	CompanyID               uuid.UUID  `json:"companyId"`
	UserID                  *uuid.UUID `json:"userId"`
	EmployeeIDNumber        string     `json:"employeeIdNumber"`
	FirstName               string     `json:"firstName"`
	LastName                string     `json:"lastName"`
	Email                   string     `json:"email"`
	Phone                   *string    `json:"phone"`
	DateOfBirth             *string    `json:"dateOfBirth"`
	Gender                  *string    `json:"gender"`
	MaritalStatus           *string    `json:"maritalStatus"`
	Address                 *string    `json:"address"`
	City                    *string    `json:"city"`
	Province                *string    `json:"province"`
	PostalCode              *string    `json:"postalCode"`
	NationalID              *string    `json:"nationalId"`
	NPWP                    *string    `json:"npwp"`
	BPJSKesehatanID         *string    `json:"bpjsKesehatanId"`
	BPJSKetenagakerjaanID   *string    `json:"bpjsKetenagakerjaanId"`
	BankName                *string    `json:"bankName"`
	BankAccountNumber       *string    `json:"bankAccountNumber"`
	BankAccountHolder       *string    `json:"bankAccountHolder"`
	EmergencyContactName    *string    `json:"emergencyContactName"`
	EmergencyContactPhone   *string    `json:"emergencyContactPhone"`
	EmergencyContactRelation *string   `json:"emergencyContactRelation"`
	EmploymentType          string     `json:"employmentType"`
	EmploymentStatus        string     `json:"employmentStatus"`
	JoinDate                string     `json:"joinDate"`
	EndDate                 *string    `json:"endDate"`
	DepartmentID            *uuid.UUID `json:"departmentId"`
	PositionID              *uuid.UUID `json:"positionId"`
	BranchID                *uuid.UUID `json:"branchId"`
	ManagerID               *uuid.UUID `json:"managerId"`
	BaseSalary              *float64   `json:"baseSalary"`
	DepartmentName          *string    `json:"departmentName"`
	PositionName            *string    `json:"positionName"`
	BranchName              *string    `json:"branchName"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

type CreateRequest struct {
	FirstName               string     `json:"firstName" binding:"required"`
	LastName                string     `json:"lastName"`
	Email                   string     `json:"email" binding:"required,email"`
	Phone                   *string    `json:"phone"`
	DateOfBirth             *string    `json:"dateOfBirth"`
	Gender                  *string    `json:"gender"`
	MaritalStatus           *string    `json:"maritalStatus"`
	Address                 *string    `json:"address"`
	City                    *string    `json:"city"`
	Province                *string    `json:"province"`
	PostalCode              *string    `json:"postalCode"`
	NationalID              *string    `json:"nationalId"`
	NPWP                    *string    `json:"npwp"`
	BPJSKesehatanID         *string    `json:"bpjsKesehatanId"`
	BPJSKetenagakerjaanID   *string    `json:"bpjsKetenagakerjaanId"`
	BankName                *string    `json:"bankName"`
	BankAccountNumber       *string    `json:"bankAccountNumber"`
	BankAccountHolder       *string    `json:"bankAccountHolder"`
	EmergencyContactName    *string    `json:"emergencyContactName"`
	EmergencyContactPhone   *string    `json:"emergencyContactPhone"`
	EmergencyContactRelation *string   `json:"emergencyContactRelation"`
	EmploymentType          string     `json:"employmentType" binding:"required,oneof=permanent contract probation intern"`
	JoinDate                string     `json:"joinDate" binding:"required"`
	DepartmentID            *uuid.UUID `json:"departmentId"`
	PositionID              *uuid.UUID `json:"positionId"`
	BranchID                *uuid.UUID `json:"branchId"`
	ManagerID               *uuid.UUID `json:"managerId"`
	BaseSalary              *float64   `json:"baseSalary"`
}

type UpdateRequest struct {
	FirstName               *string    `json:"firstName"`
	LastName                *string    `json:"lastName"`
	Email                   *string    `json:"email" binding:"omitempty,email"`
	Phone                   *string    `json:"phone"`
	DateOfBirth             *string    `json:"dateOfBirth"`
	Gender                  *string    `json:"gender"`
	MaritalStatus           *string    `json:"maritalStatus"`
	Address                 *string    `json:"address"`
	City                    *string    `json:"city"`
	Province                *string    `json:"province"`
	PostalCode              *string    `json:"postalCode"`
	NationalID              *string    `json:"nationalId"`
	NPWP                    *string    `json:"npwp"`
	BPJSKesehatanID         *string    `json:"bpjsKesehatanId"`
	BPJSKetenagakerjaanID   *string    `json:"bpjsKetenagakerjaanId"`
	BankName                *string    `json:"bankName"`
	BankAccountNumber       *string    `json:"bankAccountNumber"`
	BankAccountHolder       *string    `json:"bankAccountHolder"`
	EmergencyContactName    *string    `json:"emergencyContactName"`
	EmergencyContactPhone   *string    `json:"emergencyContactPhone"`
	EmergencyContactRelation *string   `json:"emergencyContactRelation"`
	EmploymentType          *string    `json:"employmentType" binding:"omitempty,oneof=permanent contract probation intern"`
	EmploymentStatus        *string    `json:"employmentStatus" binding:"omitempty,oneof=active resigned terminated on_leave"`
	EndDate                 *string    `json:"endDate"`
	DepartmentID            *uuid.UUID `json:"departmentId"`
	PositionID              *uuid.UUID `json:"positionId"`
	BranchID                *uuid.UUID `json:"branchId"`
	ManagerID               *uuid.UUID `json:"managerId"`
	BaseSalary              *float64   `json:"baseSalary"`
}

type ListParams struct {
	CompanyID    uuid.UUID
	Status       string
	DepartmentID *uuid.UUID
	BranchID     *uuid.UUID
	Search       string
	Page         int
	PageSize     int
}

func (s *Service) generateEmployeeID(ctx context.Context, tx *sql.Tx, companyID uuid.UUID) (string, error) {
	var prefix string
	var nextVal, padding int

	err := tx.QueryRowContext(ctx,
		`UPDATE employee_id_configs SET next_sequence = next_sequence + 1
		 WHERE company_id = $1
		 RETURNING prefix, next_sequence, padding`,
		companyID,
	).Scan(&prefix, &nextVal, &padding)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%0*d", prefix, padding, nextVal), nil
}

func (s *Service) Create(ctx context.Context, companyID uuid.UUID, req CreateRequest) (*Employee, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	empID, err := s.generateEmployeeID(ctx, tx, companyID)
	if err != nil {
		return nil, fmt.Errorf("generate employee id: %w", err)
	}

	var emp Employee
	err = tx.QueryRowContext(ctx, `
		INSERT INTO employees (
			company_id, employee_id_number, first_name, last_name, email, phone,
			date_of_birth, gender, marital_status, address, city, province, postal_code,
			national_id, npwp, bpjs_kesehatan_id, bpjs_ketenagakerjaan_id,
			bank_name, bank_account_number, bank_account_holder,
			emergency_contact_name, emergency_contact_phone, emergency_contact_relation,
			employment_type, join_date, department_id, position_id, branch_id, manager_id, base_salary
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30
		) RETURNING id, company_id, employee_id_number, first_name, last_name, email,
		  employment_type::text, employment_status::text, join_date::text, created_at, updated_at`,
		companyID, empID, req.FirstName, req.LastName, req.Email, req.Phone,
		req.DateOfBirth, req.Gender, req.MaritalStatus, req.Address, req.City, req.Province, req.PostalCode,
		req.NationalID, req.NPWP, req.BPJSKesehatanID, req.BPJSKetenagakerjaanID,
		req.BankName, req.BankAccountNumber, req.BankAccountHolder,
		req.EmergencyContactName, req.EmergencyContactPhone, req.EmergencyContactRelation,
		req.EmploymentType, req.JoinDate, req.DepartmentID, req.PositionID, req.BranchID, req.ManagerID, req.BaseSalary,
	).Scan(
		&emp.ID, &emp.CompanyID, &emp.EmployeeIDNumber, &emp.FirstName, &emp.LastName, &emp.Email,
		&emp.EmploymentType, &emp.EmploymentStatus, &emp.JoinDate, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert employee: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &emp, nil
}

func (s *Service) Get(ctx context.Context, companyID, employeeID uuid.UUID) (*Employee, error) {
	var emp Employee
	err := s.db.QueryRowContext(ctx, `
		SELECT e.id, e.company_id, e.user_id, e.employee_id_number,
			e.first_name, e.last_name, e.email, e.phone,
			e.date_of_birth::text, e.gender, e.marital_status,
			e.address, e.city, e.province, e.postal_code,
			e.national_id, e.npwp, e.bpjs_kesehatan_id, e.bpjs_ketenagakerjaan_id,
			e.bank_name, e.bank_account_number, e.bank_account_holder,
			e.emergency_contact_name, e.emergency_contact_phone, e.emergency_contact_relation,
			e.employment_type::text, e.employment_status::text,
			e.join_date::text, e.end_date::text,
			e.department_id, e.position_id, e.branch_id, e.manager_id, e.base_salary,
			d.name, p.name, b.name,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN positions p ON p.id = e.position_id
		LEFT JOIN branches b ON b.id = e.branch_id
		WHERE e.id = $1 AND e.company_id = $2`,
		employeeID, companyID,
	).Scan(
		&emp.ID, &emp.CompanyID, &emp.UserID, &emp.EmployeeIDNumber,
		&emp.FirstName, &emp.LastName, &emp.Email, &emp.Phone,
		&emp.DateOfBirth, &emp.Gender, &emp.MaritalStatus,
		&emp.Address, &emp.City, &emp.Province, &emp.PostalCode,
		&emp.NationalID, &emp.NPWP, &emp.BPJSKesehatanID, &emp.BPJSKetenagakerjaanID,
		&emp.BankName, &emp.BankAccountNumber, &emp.BankAccountHolder,
		&emp.EmergencyContactName, &emp.EmergencyContactPhone, &emp.EmergencyContactRelation,
		&emp.EmploymentType, &emp.EmploymentStatus,
		&emp.JoinDate, &emp.EndDate,
		&emp.DepartmentID, &emp.PositionID, &emp.BranchID, &emp.ManagerID, &emp.BaseSalary,
		&emp.DepartmentName, &emp.PositionName, &emp.BranchName,
		&emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

type ListResult struct {
	Employees []Employee `json:"employees"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"pageSize"`
}

func (s *Service) List(ctx context.Context, params ListParams) (*ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	baseWhere := "WHERE e.company_id = $1"
	args := []any{params.CompanyID}
	argIdx := 2

	if params.Status != "" {
		baseWhere += fmt.Sprintf(" AND e.employment_status = $%d", argIdx)
		args = append(args, params.Status)
		argIdx++
	}
	if params.DepartmentID != nil {
		baseWhere += fmt.Sprintf(" AND e.department_id = $%d", argIdx)
		args = append(args, *params.DepartmentID)
		argIdx++
	}
	if params.BranchID != nil {
		baseWhere += fmt.Sprintf(" AND e.branch_id = $%d", argIdx)
		args = append(args, *params.BranchID)
		argIdx++
	}
	if params.Search != "" {
		baseWhere += fmt.Sprintf(
			" AND (e.first_name ILIKE $%d OR e.last_name ILIKE $%d OR e.email ILIKE $%d OR e.employee_id_number ILIKE $%d)",
			argIdx, argIdx, argIdx, argIdx,
		)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	var total int
	countArgs := make([]any, len(args))
	copy(countArgs, args)
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM employees e "+baseWhere,
		countArgs...,
	).Scan(&total)
	if err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	query := fmt.Sprintf(`
		SELECT e.id, e.company_id, e.employee_id_number,
			e.first_name, e.last_name, e.email,
			e.employment_type::text, e.employment_status::text,
			e.join_date::text,
			e.department_id, e.position_id, e.branch_id,
			d.name, p.name, b.name,
			e.created_at, e.updated_at
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN positions p ON p.id = e.position_id
		LEFT JOIN branches b ON b.id = e.branch_id
		%s
		ORDER BY e.first_name, e.last_name
		LIMIT $%d OFFSET $%d`,
		baseWhere, argIdx, argIdx+1,
	)
	args = append(args, params.PageSize, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(
			&emp.ID, &emp.CompanyID, &emp.EmployeeIDNumber,
			&emp.FirstName, &emp.LastName, &emp.Email,
			&emp.EmploymentType, &emp.EmploymentStatus,
			&emp.JoinDate,
			&emp.DepartmentID, &emp.PositionID, &emp.BranchID,
			&emp.DepartmentName, &emp.PositionName, &emp.BranchName,
			&emp.CreatedAt, &emp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if employees == nil {
		employees = []Employee{}
	}

	return &ListResult{
		Employees: employees,
		Total:     total,
		Page:      params.Page,
		PageSize:  params.PageSize,
	}, nil
}

func (s *Service) Update(ctx context.Context, companyID, employeeID uuid.UUID, req UpdateRequest) (*Employee, error) {
	setClauses := []string{"updated_at = now()"}
	args := []any{companyID, employeeID}
	argIdx := 3

	addField := func(clause string, val any) {
		if val != nil {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", clause, argIdx))
			args = append(args, val)
			argIdx++
		}
	}

	addField("first_name", req.FirstName)
	addField("last_name", req.LastName)
	addField("email", req.Email)
	addField("phone", req.Phone)
	addField("date_of_birth", req.DateOfBirth)
	addField("gender", req.Gender)
	addField("marital_status", req.MaritalStatus)
	addField("address", req.Address)
	addField("city", req.City)
	addField("province", req.Province)
	addField("postal_code", req.PostalCode)
	addField("national_id", req.NationalID)
	addField("npwp", req.NPWP)
	addField("bpjs_kesehatan_id", req.BPJSKesehatanID)
	addField("bpjs_ketenagakerjaan_id", req.BPJSKetenagakerjaanID)
	addField("bank_name", req.BankName)
	addField("bank_account_number", req.BankAccountNumber)
	addField("bank_account_holder", req.BankAccountHolder)
	addField("emergency_contact_name", req.EmergencyContactName)
	addField("emergency_contact_phone", req.EmergencyContactPhone)
	addField("emergency_contact_relation", req.EmergencyContactRelation)
	addField("employment_type", req.EmploymentType)
	addField("employment_status", req.EmploymentStatus)
	addField("end_date", req.EndDate)
	addField("department_id", req.DepartmentID)
	addField("position_id", req.PositionID)
	addField("branch_id", req.BranchID)
	addField("manager_id", req.ManagerID)
	addField("base_salary", req.BaseSalary)

	query := fmt.Sprintf("UPDATE employees SET %s WHERE company_id = $1 AND id = $2",
		strings.Join(setClauses, ", "))

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return s.Get(ctx, companyID, employeeID)
}
