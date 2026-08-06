package employee

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"skillpass-server-go/internal/lib"
)

// ErrAlreadyLinked is returned when an employee already has a login account.
var ErrAlreadyLinked = errors.New("employee already has a login account")

// ErrEmailTaken is returned when the employee's email already belongs to a user.
var ErrEmailTaken = errors.New("a user account with this email already exists")

// randomPassword returns a short URL-safe temporary password.
func randomPassword() (string, error) {
	b := make([]byte, 9)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// InviteUser creates a login account for an employee (linking users.id to
// employees.user_id) so they can sign in and use HRIS self-service. Returns the
// login email and a one-time temporary password for HR to share.
//
// The email-taken check, user INSERT and employee UPDATE run in a single
// transaction; the INSERT uses ON CONFLICT (email) DO NOTHING so two
// concurrent invites for the same email can never both succeed.
func (s *Service) InviteUser(ctx context.Context, companyID, employeeID uuid.UUID) (email, tempPassword string, err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var first, last string
	var existingUserID *uuid.UUID
	err = tx.QueryRowContext(ctx,
		`SELECT email, first_name, last_name, user_id FROM employees WHERE id = $1 AND company_id = $2`,
		employeeID, companyID,
	).Scan(&email, &first, &last, &existingUserID)
	if err != nil {
		return "", "", err
	}
	if existingUserID != nil {
		return "", "", ErrAlreadyLinked
	}

	tempPassword, err = randomPassword()
	if err != nil {
		return "", "", fmt.Errorf("generate password: %w", err)
	}
	hash, err := lib.HashPassword(tempPassword)
	if err != nil {
		return "", "", fmt.Errorf("hash password: %w", err)
	}
	name := strings.TrimSpace(first + " " + last)
	if name == "" {
		name = email
	}

	// Atomic email-taken check: under a concurrent invite the loser sees no
	// inserted row here (sql.ErrNoRows) instead of racing past a pre-check.
	var userID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`INSERT INTO users (email, username, password_hash, role, name, is_verified)
		 VALUES ($1, $2, $3, 'company', $4, true)
		 ON CONFLICT (email) DO NOTHING
		 RETURNING id`,
		email, email, hash, name,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrEmailTaken
	}
	if err != nil {
		return "", "", fmt.Errorf("create user: %w", err)
	}

	if _, err = tx.ExecContext(ctx,
		`UPDATE employees SET user_id = $1, updated_at = now() WHERE id = $2 AND company_id = $3`,
		userID, employeeID, companyID,
	); err != nil {
		return "", "", fmt.Errorf("link user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", "", fmt.Errorf("commit: %w", err)
	}
	return email, tempPassword, nil
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type Employee struct {
	ID                       uuid.UUID  `json:"id"`
	CompanyID                uuid.UUID  `json:"companyId"`
	UserID                   *uuid.UUID `json:"userId,omitempty"`
	EmployeeIDNumber         string     `json:"employeeIdNumber"`
	FirstName                string     `json:"firstName"`
	LastName                 string     `json:"lastName"`
	Email                    string     `json:"email"`
	Phone                    *string    `json:"phone,omitempty"`
	DateOfBirth              *string    `json:"dateOfBirth,omitempty"`
	Gender                   *string    `json:"gender,omitempty"`
	MaritalStatus            *string    `json:"maritalStatus,omitempty"`
	Address                  *string    `json:"address,omitempty"`
	City                     *string    `json:"city,omitempty"`
	Province                 *string    `json:"province,omitempty"`
	PostalCode               *string    `json:"postalCode,omitempty"`
	NationalID               *string    `json:"nationalId,omitempty"`
	NPWP                     *string    `json:"npwp,omitempty"`
	BPJSKesehatanID          *string    `json:"bpjsKesehatanId,omitempty"`
	BPJSKetenagakerjaanID    *string    `json:"bpjsKetenagakerjaanId,omitempty"`
	BankName                 *string    `json:"bankName,omitempty"`
	BankAccountNumber        *string    `json:"bankAccountNumber,omitempty"`
	BankAccountHolder        *string    `json:"bankAccountHolder,omitempty"`
	EmergencyContactName     *string    `json:"emergencyContactName,omitempty"`
	EmergencyContactPhone    *string    `json:"emergencyContactPhone,omitempty"`
	EmergencyContactRelation *string    `json:"emergencyContactRelation,omitempty"`
	EmploymentType           string     `json:"employmentType"`
	EmploymentStatus         string     `json:"employmentStatus"`
	JoinDate                 string     `json:"joinDate"`
	EndDate                  *string    `json:"endDate,omitempty"`
	DepartmentID             *uuid.UUID `json:"departmentId,omitempty"`
	PositionID               *uuid.UUID `json:"positionId,omitempty"`
	BranchID                 *uuid.UUID `json:"branchId,omitempty"`
	ManagerID                *uuid.UUID `json:"managerId,omitempty"`
	BaseSalary               *float64   `json:"baseSalary,omitempty"`
	DepartmentName           *string    `json:"departmentName,omitempty"`
	PositionName             *string    `json:"positionName,omitempty"`
	BranchName               *string    `json:"branchName,omitempty"`
	CreatedAt                time.Time  `json:"createdAt"`
	UpdatedAt                time.Time  `json:"updatedAt"`
} //@name Employee

type CreateRequest struct {
	FirstName                string     `json:"firstName" binding:"required"`
	LastName                 string     `json:"lastName"`
	Email                    string     `json:"email" binding:"required,email"`
	Phone                    *string    `json:"phone,omitempty"`
	DateOfBirth              *string    `json:"dateOfBirth,omitempty"`
	Gender                   *string    `json:"gender,omitempty"`
	MaritalStatus            *string    `json:"maritalStatus,omitempty"`
	Address                  *string    `json:"address,omitempty"`
	City                     *string    `json:"city,omitempty"`
	Province                 *string    `json:"province,omitempty"`
	PostalCode               *string    `json:"postalCode,omitempty"`
	NationalID               *string    `json:"nationalId,omitempty"`
	NPWP                     *string    `json:"npwp,omitempty"`
	BPJSKesehatanID          *string    `json:"bpjsKesehatanId,omitempty"`
	BPJSKetenagakerjaanID    *string    `json:"bpjsKetenagakerjaanId,omitempty"`
	BankName                 *string    `json:"bankName,omitempty"`
	BankAccountNumber        *string    `json:"bankAccountNumber,omitempty"`
	BankAccountHolder        *string    `json:"bankAccountHolder,omitempty"`
	EmergencyContactName     *string    `json:"emergencyContactName,omitempty"`
	EmergencyContactPhone    *string    `json:"emergencyContactPhone,omitempty"`
	EmergencyContactRelation *string    `json:"emergencyContactRelation,omitempty"`
	EmploymentType           string     `json:"employmentType" binding:"required,oneof=permanent contract probation intern"`
	JoinDate                 string     `json:"joinDate" binding:"required"`
	DepartmentID             *uuid.UUID `json:"departmentId,omitempty"`
	PositionID               *uuid.UUID `json:"positionId,omitempty"`
	BranchID                 *uuid.UUID `json:"branchId,omitempty"`
	ManagerID                *uuid.UUID `json:"managerId,omitempty"`
	BaseSalary               *float64   `json:"baseSalary,omitempty"`
} //@name CreateRequest

type UpdateRequest struct {
	FirstName                *string    `json:"firstName,omitempty"`
	LastName                 *string    `json:"lastName,omitempty"`
	Email                    *string    `json:"email,omitempty" binding:"omitempty,email"`
	Phone                    *string    `json:"phone,omitempty"`
	DateOfBirth              *string    `json:"dateOfBirth,omitempty"`
	Gender                   *string    `json:"gender,omitempty"`
	MaritalStatus            *string    `json:"maritalStatus,omitempty"`
	Address                  *string    `json:"address,omitempty"`
	City                     *string    `json:"city,omitempty"`
	Province                 *string    `json:"province,omitempty"`
	PostalCode               *string    `json:"postalCode,omitempty"`
	NationalID               *string    `json:"nationalId,omitempty"`
	NPWP                     *string    `json:"npwp,omitempty"`
	BPJSKesehatanID          *string    `json:"bpjsKesehatanId,omitempty"`
	BPJSKetenagakerjaanID    *string    `json:"bpjsKetenagakerjaanId,omitempty"`
	BankName                 *string    `json:"bankName,omitempty"`
	BankAccountNumber        *string    `json:"bankAccountNumber,omitempty"`
	BankAccountHolder        *string    `json:"bankAccountHolder,omitempty"`
	EmergencyContactName     *string    `json:"emergencyContactName,omitempty"`
	EmergencyContactPhone    *string    `json:"emergencyContactPhone,omitempty"`
	EmergencyContactRelation *string    `json:"emergencyContactRelation,omitempty"`
	EmploymentType           *string    `json:"employmentType,omitempty" binding:"omitempty,oneof=permanent contract probation intern"`
	EmploymentStatus         *string    `json:"employmentStatus,omitempty" binding:"omitempty,oneof=active resigned terminated on_leave"`
	EndDate                  *string    `json:"endDate,omitempty"`
	DepartmentID             *uuid.UUID `json:"departmentId,omitempty"`
	PositionID               *uuid.UUID `json:"positionId,omitempty"`
	BranchID                 *uuid.UUID `json:"branchId,omitempty"`
	ManagerID                *uuid.UUID `json:"managerId,omitempty"`
	BaseSalary               *float64   `json:"baseSalary,omitempty"`
} //@name UpdateRequest

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
	Employees []Employee `json:"employees,omitempty"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"pageSize"`
} //@name ListResult

func (s *Service) List(ctx context.Context, params ListParams) (*ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	// Clamp to a sane cap instead of silently falling back to 20 when the
	// caller asks for a larger page (e.g. CSV export requesting pageSize 1000).
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
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
		// Match the pg_trgm index expression from migration 000033/000034
		// (COALESCE so NULL columns don't drop the row from the index).
		baseWhere += fmt.Sprintf(
			" AND (COALESCE(e.first_name,'') || ' ' || COALESCE(e.last_name,'') || ' ' || COALESCE(e.email,'') || ' ' || COALESCE(e.employee_id_number,'')) ILIKE $%d",
			argIdx,
		)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	var total int
	// Copy args for the COUNT query because the main query below appends
	// pagination parameters (LIMIT, OFFSET) to the same args slice.
	// Without the copy, those extra params would be sent to COUNT too.
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
		if val == nil {
			return
		}
		// typed nil pointer — interface holds type info even when pointer is nil
		switch v := val.(type) {
		case *string:
			if v == nil {
				return
			}
		case *uuid.UUID:
			if v == nil {
				return
			}
		case *float64:
			if v == nil {
				return
			}
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", clause, argIdx))
		args = append(args, val)
		argIdx++
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
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return nil, sql.ErrNoRows
	}

	return s.Get(ctx, companyID, employeeID)
}
