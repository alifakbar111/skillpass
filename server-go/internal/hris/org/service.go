package org

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

// ============================================================
// Branch
// ============================================================

type Branch struct {
	ID                  uuid.UUID  `json:"id"`
	CompanyID           uuid.UUID  `json:"companyId"`
	Name                string     `json:"name"`
	BranchType          string     `json:"branchType"`
	ParentBranchID      *uuid.UUID `json:"parentBranchId"`
	Address             *string    `json:"address"`
	City                *string    `json:"city"`
	Province            *string    `json:"province"`
	Latitude            *float64   `json:"latitude"`
	Longitude           *float64   `json:"longitude"`
	GeofenceRadiusMeters int       `json:"geofenceRadiusMeters"`
	IsActive            bool       `json:"isActive"`
	CreatedAt           time.Time  `json:"createdAt"`
}

type CreateBranchRequest struct {
	Name                string     `json:"name" binding:"required"`
	BranchType          string     `json:"branchType" binding:"required,oneof=head_office regional branch"`
	ParentBranchID      *uuid.UUID `json:"parentBranchId"`
	Address             *string    `json:"address"`
	City                *string    `json:"city"`
	Province            *string    `json:"province"`
	Latitude            *float64   `json:"latitude"`
	Longitude           *float64   `json:"longitude"`
	GeofenceRadiusMeters *int      `json:"geofenceRadiusMeters"`
}

type UpdateBranchRequest struct {
	Name                *string    `json:"name"`
	BranchType          *string    `json:"branchType" binding:"omitempty,oneof=head_office regional branch"`
	ParentBranchID      *uuid.UUID `json:"parentBranchId"`
	Address             *string    `json:"address"`
	City                *string    `json:"city"`
	Province            *string    `json:"province"`
	Latitude            *float64   `json:"latitude"`
	Longitude           *float64   `json:"longitude"`
	GeofenceRadiusMeters *int      `json:"geofenceRadiusMeters"`
	IsActive            *bool      `json:"isActive"`
}

func (s *Service) CreateBranch(ctx context.Context, companyID uuid.UUID, req CreateBranchRequest) (*Branch, error) {
	radius := 200
	if req.GeofenceRadiusMeters != nil {
		radius = *req.GeofenceRadiusMeters
	}

	var b Branch
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO branches (company_id, name, branch_type, parent_branch_id, address, city, province, latitude, longitude, geofence_radius_meters)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, company_id, name, branch_type, parent_branch_id, address, city, province, latitude, longitude, geofence_radius_meters, is_active, created_at`,
		companyID, req.Name, req.BranchType, req.ParentBranchID, req.Address, req.City, req.Province, req.Latitude, req.Longitude, radius,
	).Scan(&b.ID, &b.CompanyID, &b.Name, &b.BranchType, &b.ParentBranchID, &b.Address, &b.City, &b.Province, &b.Latitude, &b.Longitude, &b.GeofenceRadiusMeters, &b.IsActive, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) ListBranches(ctx context.Context, companyID uuid.UUID) ([]Branch, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company_id, name, branch_type, parent_branch_id, address, city, province, latitude, longitude, geofence_radius_meters, is_active, created_at
		FROM branches WHERE company_id = $1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []Branch
	for rows.Next() {
		var b Branch
		if err := rows.Scan(&b.ID, &b.CompanyID, &b.Name, &b.BranchType, &b.ParentBranchID, &b.Address, &b.City, &b.Province, &b.Latitude, &b.Longitude, &b.GeofenceRadiusMeters, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	if branches == nil {
		branches = []Branch{}
	}
	return branches, rows.Err()
}

func (s *Service) GetBranch(ctx context.Context, companyID, branchID uuid.UUID) (*Branch, error) {
	var b Branch
	err := s.db.QueryRowContext(ctx, `
		SELECT id, company_id, name, branch_type, parent_branch_id, address, city, province, latitude, longitude, geofence_radius_meters, is_active, created_at
		FROM branches WHERE id = $1 AND company_id = $2`, branchID, companyID,
	).Scan(&b.ID, &b.CompanyID, &b.Name, &b.BranchType, &b.ParentBranchID, &b.Address, &b.City, &b.Province, &b.Latitude, &b.Longitude, &b.GeofenceRadiusMeters, &b.IsActive, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) UpdateBranch(ctx context.Context, companyID, branchID uuid.UUID, req UpdateBranchRequest) (*Branch, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE branches SET
			name = COALESCE($3, name),
			branch_type = COALESCE($4, branch_type),
			parent_branch_id = COALESCE($5, parent_branch_id),
			address = COALESCE($6, address),
			city = COALESCE($7, city),
			province = COALESCE($8, province),
			latitude = COALESCE($9, latitude),
			longitude = COALESCE($10, longitude),
			geofence_radius_meters = COALESCE($11, geofence_radius_meters),
			is_active = COALESCE($12, is_active)
		WHERE id = $1 AND company_id = $2`,
		branchID, companyID, req.Name, req.BranchType, req.ParentBranchID, req.Address, req.City, req.Province, req.Latitude, req.Longitude, req.GeofenceRadiusMeters, req.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return s.GetBranch(ctx, companyID, branchID)
}

func (s *Service) DeleteBranch(ctx context.Context, companyID, branchID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE branches SET is_active = FALSE WHERE id = $1 AND company_id = $2`,
		branchID, companyID)
	return err
}

// ============================================================
// Department
// ============================================================

type Department struct {
	ID                 uuid.UUID  `json:"id"`
	CompanyID          uuid.UUID  `json:"companyId"`
	Name               string     `json:"name"`
	ParentDepartmentID *uuid.UUID `json:"parentDepartmentId"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type CreateDepartmentRequest struct {
	Name               string     `json:"name" binding:"required"`
	ParentDepartmentID *uuid.UUID `json:"parentDepartmentId"`
}

type UpdateDepartmentRequest struct {
	Name               *string    `json:"name"`
	ParentDepartmentID *uuid.UUID `json:"parentDepartmentId"`
}

func (s *Service) CreateDepartment(ctx context.Context, companyID uuid.UUID, req CreateDepartmentRequest) (*Department, error) {
	var d Department
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO departments (company_id, name, parent_department_id)
		VALUES ($1, $2, $3)
		RETURNING id, company_id, name, parent_department_id, created_at`,
		companyID, req.Name, req.ParentDepartmentID,
	).Scan(&d.ID, &d.CompanyID, &d.Name, &d.ParentDepartmentID, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Service) ListDepartments(ctx context.Context, companyID uuid.UUID) ([]Department, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, parent_department_id, created_at FROM departments WHERE company_id = $1 ORDER BY name`,
		companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depts []Department
	for rows.Next() {
		var d Department
		if err := rows.Scan(&d.ID, &d.CompanyID, &d.Name, &d.ParentDepartmentID, &d.CreatedAt); err != nil {
			return nil, err
		}
		depts = append(depts, d)
	}
	if depts == nil {
		depts = []Department{}
	}
	return depts, rows.Err()
}

func (s *Service) UpdateDepartment(ctx context.Context, companyID, deptID uuid.UUID, req UpdateDepartmentRequest) (*Department, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE departments SET
			name = COALESCE($3, name),
			parent_department_id = COALESCE($4, parent_department_id)
		WHERE id = $1 AND company_id = $2`,
		deptID, companyID, req.Name, req.ParentDepartmentID,
	)
	if err != nil {
		return nil, err
	}
	var d Department
	err = s.db.QueryRowContext(ctx,
		`SELECT id, company_id, name, parent_department_id, created_at FROM departments WHERE id = $1 AND company_id = $2`,
		deptID, companyID,
	).Scan(&d.ID, &d.CompanyID, &d.Name, &d.ParentDepartmentID, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Service) DeleteDepartment(ctx context.Context, companyID, deptID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM departments WHERE id = $1 AND company_id = $2`, deptID, companyID)
	return err
}

// ============================================================
// Position
// ============================================================

type Position struct {
	ID           uuid.UUID  `json:"id"`
	CompanyID    uuid.UUID  `json:"companyId"`
	Name         string     `json:"name"`
	DepartmentID *uuid.UUID `json:"departmentId"`
	Level        string     `json:"level"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type CreatePositionRequest struct {
	Name         string     `json:"name" binding:"required"`
	DepartmentID *uuid.UUID `json:"departmentId"`
	Level        string     `json:"level" binding:"required,oneof=staff supervisor manager director"`
}

type UpdatePositionRequest struct {
	Name         *string    `json:"name"`
	DepartmentID *uuid.UUID `json:"departmentId"`
	Level        *string    `json:"level" binding:"omitempty,oneof=staff supervisor manager director"`
}

func (s *Service) CreatePosition(ctx context.Context, companyID uuid.UUID, req CreatePositionRequest) (*Position, error) {
	var p Position
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO positions (company_id, name, department_id, level)
		VALUES ($1, $2, $3, $4)
		RETURNING id, company_id, name, department_id, level::text, created_at`,
		companyID, req.Name, req.DepartmentID, req.Level,
	).Scan(&p.ID, &p.CompanyID, &p.Name, &p.DepartmentID, &p.Level, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) ListPositions(ctx context.Context, companyID uuid.UUID) ([]Position, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, department_id, level::text, created_at FROM positions WHERE company_id = $1 ORDER BY name`,
		companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []Position
	for rows.Next() {
		var p Position
		if err := rows.Scan(&p.ID, &p.CompanyID, &p.Name, &p.DepartmentID, &p.Level, &p.CreatedAt); err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	if positions == nil {
		positions = []Position{}
	}
	return positions, rows.Err()
}

func (s *Service) UpdatePosition(ctx context.Context, companyID, posID uuid.UUID, req UpdatePositionRequest) (*Position, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE positions SET
			name = COALESCE($3, name),
			department_id = COALESCE($4, department_id),
			level = COALESCE($5, level)
		WHERE id = $1 AND company_id = $2`,
		posID, companyID, req.Name, req.DepartmentID, req.Level,
	)
	if err != nil {
		return nil, err
	}
	var p Position
	err = s.db.QueryRowContext(ctx,
		`SELECT id, company_id, name, department_id, level::text, created_at FROM positions WHERE id = $1 AND company_id = $2`,
		posID, companyID,
	).Scan(&p.ID, &p.CompanyID, &p.Name, &p.DepartmentID, &p.Level, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) DeletePosition(ctx context.Context, companyID, posID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM positions WHERE id = $1 AND company_id = $2`, posID, companyID)
	return err
}

// ============================================================
// Org Tree
// ============================================================

type OrgNode struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	ParentID     *uuid.UUID `json:"parentId"`
	Level        *string    `json:"level"`
	EmployeeCount int       `json:"employeeCount,omitempty"`
	Children     []OrgNode  `json:"children,omitempty"`
}

func (s *Service) GetOrgTree(ctx context.Context, companyID uuid.UUID) ([]OrgNode, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT d.id, d.name, d.parent_department_id, COUNT(e.id) as emp_count
		FROM departments d
		LEFT JOIN employees e ON e.department_id = d.id AND e.employment_status = 'active'
		WHERE d.company_id = $1
		GROUP BY d.id, d.name, d.parent_department_id
		ORDER BY d.name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flatNodes []OrgNode
	for rows.Next() {
		var n OrgNode
		if err := rows.Scan(&n.ID, &n.Name, &n.ParentID, &n.EmployeeCount); err != nil {
			return nil, err
		}
		n.Type = "department"
		flatNodes = append(flatNodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buildTree(flatNodes), nil
}

// ============================================================
// Working Calendars
// ============================================================

type WorkingCalendar struct {
	ID              uuid.UUID  `json:"id"`
	CompanyID       uuid.UUID  `json:"companyId"`
	BranchID        *uuid.UUID `json:"branchId"`
	Year            int        `json:"year"`
	DefaultWorkDays []int      `json:"defaultWorkDays"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type CreateCalendarRequest struct {
	BranchID        *uuid.UUID `json:"branchId"`
	Year            int        `json:"year" binding:"required"`
	DefaultWorkDays []int      `json:"defaultWorkDays" binding:"required"`
}

type UpdateCalendarRequest struct {
	DefaultWorkDays []int `json:"defaultWorkDays" binding:"required"`
}

func (s *Service) CreateCalendar(ctx context.Context, companyID uuid.UUID, req CreateCalendarRequest) (*WorkingCalendar, error) {
	var wc WorkingCalendar
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO working_calendars (company_id, branch_id, year, default_work_days)
		VALUES ($1, $2, $3, $4)
		RETURNING id, company_id, branch_id, year, default_work_days, created_at`,
		companyID, req.BranchID, req.Year, pqIntArray(req.DefaultWorkDays),
	).Scan(&wc.ID, &wc.CompanyID, &wc.BranchID, &wc.Year, pqIntArrayScanner(&wc.DefaultWorkDays), &wc.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &wc, nil
}

func (s *Service) ListCalendars(ctx context.Context, companyID uuid.UUID, year *int) ([]WorkingCalendar, error) {
	query := `SELECT id, company_id, branch_id, year, default_work_days, created_at
		FROM working_calendars WHERE company_id = $1`
	args := []interface{}{companyID}
	if year != nil {
		query += ` AND year = $2`
		args = append(args, *year)
	}
	query += ` ORDER BY year DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calendars []WorkingCalendar
	for rows.Next() {
		var wc WorkingCalendar
		if err := rows.Scan(&wc.ID, &wc.CompanyID, &wc.BranchID, &wc.Year, pqIntArrayScanner(&wc.DefaultWorkDays), &wc.CreatedAt); err != nil {
			return nil, err
		}
		calendars = append(calendars, wc)
	}
	if calendars == nil {
		calendars = []WorkingCalendar{}
	}
	return calendars, rows.Err()
}

func (s *Service) UpdateCalendar(ctx context.Context, companyID, calendarID uuid.UUID, req UpdateCalendarRequest) (*WorkingCalendar, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE working_calendars SET default_work_days = $3
		WHERE id = $1 AND company_id = $2`,
		calendarID, companyID, pqIntArray(req.DefaultWorkDays),
	)
	if err != nil {
		return nil, err
	}

	var wc WorkingCalendar
	err = s.db.QueryRowContext(ctx, `
		SELECT id, company_id, branch_id, year, default_work_days, created_at
		FROM working_calendars WHERE id = $1 AND company_id = $2`,
		calendarID, companyID,
	).Scan(&wc.ID, &wc.CompanyID, &wc.BranchID, &wc.Year, pqIntArrayScanner(&wc.DefaultWorkDays), &wc.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &wc, nil
}

func (s *Service) DeleteCalendar(ctx context.Context, companyID, calendarID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM working_calendars WHERE id = $1 AND company_id = $2`,
		calendarID, companyID)
	return err
}

// ============================================================
// Enhanced Org Chart (employee-level recursive tree)
// ============================================================

type OrgChartNode struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	PositionName *string        `json:"positionName"`
	Level        *string        `json:"level"`
	DepartmentID *uuid.UUID     `json:"departmentId"`
	ManagerID    *uuid.UUID     `json:"managerId"`
	Children     []OrgChartNode `json:"children"`
}

func (s *Service) GetOrgChart(ctx context.Context, companyID uuid.UUID) ([]OrgChartNode, error) {
	rows, err := s.db.QueryContext(ctx, `
		WITH RECURSIVE org_tree AS (
			SELECT e.id, e.first_name || ' ' || e.last_name AS name, e.manager_id,
				   e.department_id, p.name AS position_name, p.level::text AS level
			FROM employees e
			LEFT JOIN positions p ON p.id = e.position_id
			WHERE e.company_id = $1 AND e.employment_status = 'active' AND e.manager_id IS NULL
			UNION ALL
			SELECT e.id, e.first_name || ' ' || e.last_name, e.manager_id,
				   e.department_id, p.name, p.level::text
			FROM employees e
			LEFT JOIN positions p ON p.id = e.position_id
			JOIN org_tree ot ON e.manager_id = ot.id
			WHERE e.employment_status = 'active'
		)
		SELECT id, name, manager_id, department_id, position_name, level FROM org_tree`,
		companyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flatNodes []OrgChartNode
	for rows.Next() {
		var n OrgChartNode
		if err := rows.Scan(&n.ID, &n.Name, &n.ManagerID, &n.DepartmentID, &n.PositionName, &n.Level); err != nil {
			return nil, err
		}
		flatNodes = append(flatNodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buildChartTree(flatNodes), nil
}

func buildChartTree(nodes []OrgChartNode) []OrgChartNode {
	nodeMap := make(map[uuid.UUID]*OrgChartNode)
	for i := range nodes {
		nodes[i].Children = []OrgChartNode{}
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	var roots []OrgChartNode
	for i := range nodes {
		if nodes[i].ManagerID == nil {
			roots = append(roots, nodes[i])
		} else if parent, ok := nodeMap[*nodes[i].ManagerID]; ok {
			parent.Children = append(parent.Children, nodes[i])
		} else {
			roots = append(roots, nodes[i])
		}
	}
	if roots == nil {
		roots = []OrgChartNode{}
	}
	return roots
}

// ============================================================
// Int array helpers for PostgreSQL
// ============================================================

type pqIntArray []int

func (a pqIntArray) Value() (interface{}, error) {
	if a == nil {
		return "{}", nil
	}
	s := "{"
	for i, v := range a {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%d", v)
	}
	s += "}"
	return s, nil
}

type intArrayScanner struct {
	dest *[]int
}

func pqIntArrayScanner(dest *[]int) *intArrayScanner {
	return &intArrayScanner{dest: dest}
}

func (s *intArrayScanner) Scan(src interface{}) error {
	if src == nil {
		*s.dest = []int{}
		return nil
	}
	var str string
	switch v := src.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("unsupported type for int array: %T", src)
	}
	// Parse PostgreSQL array format: {1,2,3,4,5}
	str = str[1 : len(str)-1] // strip braces
	if str == "" {
		*s.dest = []int{}
		return nil
	}
	parts := splitComma(str)
	result := make([]int, len(parts))
	for i, p := range parts {
		var n int
		_, err := fmt.Sscanf(p, "%d", &n)
		if err != nil {
			return err
		}
		result[i] = n
	}
	*s.dest = result
	return nil
}

func splitComma(s string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	return result
}

func buildTree(nodes []OrgNode) []OrgNode {
	nodeMap := make(map[uuid.UUID]*OrgNode)
	for i := range nodes {
		nodes[i].Children = []OrgNode{}
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	var roots []OrgNode
	for i := range nodes {
		if nodes[i].ParentID == nil {
			roots = append(roots, nodes[i])
		} else if parent, ok := nodeMap[*nodes[i].ParentID]; ok {
			parent.Children = append(parent.Children, nodes[i])
		} else {
			roots = append(roots, nodes[i])
		}
	}
	if roots == nil {
		roots = []OrgNode{}
	}
	return roots
}
