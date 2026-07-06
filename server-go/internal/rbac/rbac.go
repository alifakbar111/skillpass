package rbac

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type EmployeeInfo struct {
	EmployeeID uuid.UUID
	CompanyID  uuid.UUID
}

func (s *Service) GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (*EmployeeInfo, error) {
	var info EmployeeInfo
	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_id FROM employees WHERE user_id = $1 AND employment_status = 'active' LIMIT 1`,
		userID,
	).Scan(&info.EmployeeID, &info.CompanyID)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *Service) HasPermission(ctx context.Context, employeeID uuid.UUID, permCode string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM employee_roles er
			JOIN role_permissions rp ON rp.role_id = er.role_id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE er.employee_id = $1 AND p.code = $2
		)`,
		employeeID, permCode,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// TODO: Cache permission check per-request or per-session to avoid DB hit on every HRIS request
func (s *Service) HasAnyPermission(ctx context.Context, employeeID uuid.UUID, permCodes []string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM employee_roles er
			JOIN role_permissions rp ON rp.role_id = er.role_id
			JOIN permissions p ON p.id = rp.permission_id
			WHERE er.employee_id = $1 AND p.code = ANY($2)
		)`,
		employeeID, permCodes,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsSystem    bool      `json:"isSystem"`
}

func (s *Service) GetEmployeeRoles(ctx context.Context, employeeID uuid.UUID) ([]Role, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.name, r.description, r.is_system
		 FROM hris_roles r
		 JOIN employee_roles er ON er.role_id = r.id
		 WHERE er.employee_id = $1
		 ORDER BY r.name`,
		employeeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func (s *Service) GetEmployeePermissions(ctx context.Context, employeeID uuid.UUID) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT p.code
		 FROM permissions p
		 JOIN role_permissions rp ON rp.permission_id = p.id
		 JOIN employee_roles er ON er.role_id = rp.role_id
		 WHERE er.employee_id = $1
		 ORDER BY p.code`,
		employeeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, rows.Err()
}

func (s *Service) AssignRole(ctx context.Context, companyID, employeeID, roleID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO employee_roles (employee_id, role_id)
		 SELECT $2, $3
		 FROM employees e, hris_roles r
		 WHERE e.id = $2 AND e.company_id = $1
		   AND r.id = $3 AND r.company_id = $1
		 ON CONFLICT DO NOTHING`,
		companyID, employeeID, roleID,
	)
	return err
}

func (s *Service) RemoveRole(ctx context.Context, companyID, employeeID, roleID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM employee_roles er
		 USING employees e, hris_roles r
		 WHERE er.employee_id = e.id AND er.role_id = r.id
		   AND e.id = $2 AND e.company_id = $1
		   AND r.id = $3 AND r.company_id = $1`,
		companyID, employeeID, roleID,
	)
	return err
}

func (s *Service) ListRoles(ctx context.Context, companyID uuid.UUID) ([]Role, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, is_system FROM hris_roles WHERE company_id = $1 ORDER BY is_system DESC, name`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func (s *Service) EnsureCompanyRoles(ctx context.Context, companyID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO hris_roles (company_id, name, description, is_system) VALUES
			($1, 'Company Admin', 'Full access to all modules and company settings', true),
			($1, 'HR Admin', 'All HRIS modules; cannot access billing', true),
			($1, 'Payroll Admin', 'Payroll run, payslip, tax reports', true),
			($1, 'Manager', 'Team attendance, leave approval, KPI review for direct reports', true),
			($1, 'Employee', 'Own attendance, payslip, leave, KPI', true),
			($1, 'Recruiter', 'Job posting, candidate pipeline, ATS', true),
			($1, 'Director', 'Executive dashboard, analytics, read-only', true),
			($1, 'Auditor', 'Full read access, no write, audit log', true)
		ON CONFLICT (company_id, name) DO NOTHING
	`, companyID)
	if err != nil {
		return err
	}

	return s.seedRolePermissions(ctx, companyID)
}

// Permission represents a single permission record.
type Permission struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Module      string `json:"module"`
	Description string `json:"description"`
}

func (s *Service) ListPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, code, module, description FROM permissions ORDER BY module, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.Code, &p.Module, &p.Description); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, rows.Err()
}

func (s *Service) CreateRole(ctx context.Context, companyID uuid.UUID, name string, description *string) (*Role, error) {
	if name == "" {
		return nil, fmt.Errorf("role name cannot be empty")
	}

	var r Role
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO hris_roles (company_id, name, description, is_system)
		 VALUES ($1, $2, $3, false)
		 RETURNING id, name, description, is_system`,
		companyID, name, description,
	).Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem)
	if err != nil {
		if isPGUniqueViolation(err) {
			return nil, fmt.Errorf("role %q already exists in this company", name)
		}
		return nil, fmt.Errorf("insert role: %w", err)
	}
	return &r, nil
}

func (s *Service) UpdateRole(ctx context.Context, companyID uuid.UUID, roleID uuid.UUID, name string, description *string) (*Role, error) {
	if name == "" {
		return nil, fmt.Errorf("role name cannot be empty")
	}

	// Check if it's a system role — system roles cannot be modified
	var isSystem bool
	err := s.db.QueryRowContext(ctx,
		`SELECT is_system FROM hris_roles WHERE id = $1 AND company_id = $2`,
		roleID, companyID,
	).Scan(&isSystem)
	if err != nil {
		return nil, fmt.Errorf("role not found")
	}
	if isSystem {
		return nil, fmt.Errorf("cannot modify system role")
	}

	var r Role
	err = s.db.QueryRowContext(ctx,
		`UPDATE hris_roles SET name = $1, description = $2 WHERE id = $3 AND company_id = $4
		 RETURNING id, name, description, is_system`,
		name, description, roleID, companyID,
	).Scan(&r.ID, &r.Name, &r.Description, &r.IsSystem)
	if err != nil {
		if isPGUniqueViolation(err) {
			return nil, fmt.Errorf("role %q already exists in this company", name)
		}
		return nil, fmt.Errorf("update role: %w", err)
	}
	return &r, nil
}

func (s *Service) DeleteRole(ctx context.Context, companyID uuid.UUID, roleID uuid.UUID) error {
	// Check if it's a system role — system roles cannot be deleted
	var isSystem bool
	err := s.db.QueryRowContext(ctx,
		`SELECT is_system FROM hris_roles WHERE id = $1 AND company_id = $2`,
		roleID, companyID,
	).Scan(&isSystem)
	if err != nil {
		return fmt.Errorf("role not found")
	}
	if isSystem {
		return fmt.Errorf("cannot delete system role")
	}

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM hris_roles WHERE id = $1 AND company_id = $2`,
		roleID, companyID,
	)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("role not found")
	}
	return nil
}

func (s *Service) SetRolePermissions(ctx context.Context, companyID uuid.UUID, roleID uuid.UUID, permissionIDs []string) error {
	// Verify the role exists and belongs to this company
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM hris_roles WHERE id = $1 AND company_id = $2)`,
		roleID, companyID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("verify role: %w", err)
	}
	if !exists {
		return fmt.Errorf("role not found")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Remove all existing permissions for this role
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM role_permissions WHERE role_id = $1`, roleID,
	); err != nil {
		return fmt.Errorf("clear permissions: %w", err)
	}

	// Insert new permissions
	if len(permissionIDs) > 0 {
		for _, pid := range permissionIDs {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				roleID, pid,
			); err != nil {
				return fmt.Errorf("assign permission %s: %w", pid, err)
			}
		}
	}

	return tx.Commit()
}

func isPGUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint"))
}

func (s *Service) seedRolePermissions(ctx context.Context, companyID uuid.UUID) error {
	type rolePerm struct {
		role string
		perm string
	}
	var pairs []rolePerm
	for roleName, perms := range map[string][]string{
		"Company Admin": {
			"employee.view", "employee.create", "employee.update", "employee.delete",
			"attendance.view", "attendance.manage", "attendance.clock", "attendance.approve", "attendance.export",
			"leave.view", "leave.request", "leave.approve", "leave.manage",
			"payroll.view", "payroll.run", "payroll.approve", "payroll.manage",
			"performance.view", "performance.manage", "performance.review",
			"ats.view", "ats.manage", "ats.scorecard",
			"analytics.view", "analytics.view_exec", "analytics.export",
			"documents.view", "documents.manage",
			"org.view", "org.manage", "roles.manage", "settings.manage",
		},
		"HR Admin": {
			"employee.view", "employee.create", "employee.update", "employee.delete",
			"attendance.view", "attendance.manage", "attendance.clock", "attendance.approve", "attendance.export",
			"leave.view", "leave.request", "leave.approve", "leave.manage",
			"payroll.view",
			"performance.view", "performance.manage", "performance.review",
			"ats.view",
			"analytics.view", "analytics.export",
			"documents.view", "documents.manage",
			"org.view", "org.manage",
		},
		"Payroll Admin": {
			"employee.view",
			"attendance.view",
			"leave.view",
			"payroll.view", "payroll.run", "payroll.approve", "payroll.manage",
			"payroll.view_self",
		},
		"Manager": {
			"employee.view_team",
			"attendance.view_team", "attendance.clock", "attendance.approve",
			"leave.view_team", "leave.request", "leave.approve",
			"performance.view_team", "performance.review",
			"ats.scorecard",
			"analytics.view_team",
			"documents.view_team",
			"org.view",
			"payroll.view_self",
		},
		"Employee": {
			"employee.view_self",
			"attendance.view_self", "attendance.clock",
			"leave.request",
			"payroll.view_self",
			"performance.view_self",
			"documents.view_self",
			"org.view",
		},
		"Recruiter": {
			"ats.view", "ats.manage", "ats.scorecard",
			"analytics.view_team",
			"documents.view_self",
			"org.view",
		},
		"Director": {
			"employee.view",
			"attendance.view",
			"leave.view",
			"payroll.view",
			"performance.view",
			"analytics.view", "analytics.view_exec",
			"documents.view",
			"org.view",
		},
		"Auditor": {
			"employee.view",
			"attendance.view",
			"leave.view",
			"payroll.view",
			"performance.view",
			"analytics.view",
			"documents.view",
			"org.view",
		},
	} {
		for _, perm := range perms {
			pairs = append(pairs, rolePerm{role: roleName, perm: perm})
		}
	}

	if len(pairs) == 0 {
		return nil
	}

	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id
		FROM hris_roles r
		CROSS JOIN permissions p
		WHERE r.company_id = $1 AND r.is_system = true
		  AND (r.name, p.code) IN (`
	args := []any{companyID}
	argIdx := 2
	for i, pair := range pairs {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("($%d,$%d)", argIdx, argIdx+1)
		args = append(args, pair.role, pair.perm)
		argIdx += 2
	}
	query += `)
		ON CONFLICT DO NOTHING`

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}
