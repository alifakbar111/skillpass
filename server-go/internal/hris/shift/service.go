package shift

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type ShiftTemplate struct {
	ID                    uuid.UUID `json:"id"`
	CompanyID             uuid.UUID `json:"companyId"`
	Name                  string    `json:"name"`
	StartTime             string    `json:"startTime"`
	EndTime               string    `json:"endTime"`
	BreakDurationMinutes  int       `json:"breakDurationMinutes"`
	LateToleranceMinutes  int       `json:"lateToleranceMinutes"`
	OvertimeMultiplier    float64   `json:"overtimeMultiplier"`
	ApplicableDays        []int64   `json:"applicableDays"`
	IsDefault             bool      `json:"isDefault"`
	CreatedAt             time.Time `json:"createdAt"`
}

type EmployeeShift struct {
	ID            uuid.UUID  `json:"id"`
	EmployeeID    uuid.UUID  `json:"employeeId"`
	ShiftID       uuid.UUID  `json:"shiftId"`
	EffectiveDate string     `json:"effectiveDate"`
	EndDate       *string    `json:"endDate"`
	ShiftName     string     `json:"shiftName"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func (s *Service) ListTemplates(ctx context.Context, companyID uuid.UUID) ([]ShiftTemplate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, start_time::text, end_time::text,
		        break_duration_minutes, late_tolerance_minutes, overtime_multiplier,
		        applicable_days, is_default, created_at
		 FROM shift_templates WHERE company_id = $1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ShiftTemplate
	for rows.Next() {
		var t ShiftTemplate
		if err := rows.Scan(&t.ID, &t.CompanyID, &t.Name, &t.StartTime, &t.EndTime,
			&t.BreakDurationMinutes, &t.LateToleranceMinutes, &t.OvertimeMultiplier,
			pq.Array(&t.ApplicableDays), &t.IsDefault, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (s *Service) CreateTemplate(ctx context.Context, companyID uuid.UUID, t *ShiftTemplate) error {
	return s.db.QueryRowContext(ctx,
		`INSERT INTO shift_templates (company_id, name, start_time, end_time,
		  break_duration_minutes, late_tolerance_minutes, overtime_multiplier,
		  applicable_days, is_default)
		 VALUES ($1, $2, $3::time, $4::time, $5, $6, $7, $8, $9)
		 RETURNING id, created_at`,
		companyID, t.Name, t.StartTime, t.EndTime,
		t.BreakDurationMinutes, t.LateToleranceMinutes, t.OvertimeMultiplier,
		pq.Array(t.ApplicableDays), t.IsDefault,
	).Scan(&t.ID, &t.CreatedAt)
}

func (s *Service) UpdateTemplate(ctx context.Context, companyID, id uuid.UUID, t *ShiftTemplate) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE shift_templates SET name=$1, start_time=$2::time, end_time=$3::time,
		  break_duration_minutes=$4, late_tolerance_minutes=$5, overtime_multiplier=$6,
		  applicable_days=$7, is_default=$8
		 WHERE id=$9 AND company_id=$10`,
		t.Name, t.StartTime, t.EndTime,
		t.BreakDurationMinutes, t.LateToleranceMinutes, t.OvertimeMultiplier,
		pq.Array(t.ApplicableDays), t.IsDefault, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Service) DeleteTemplate(ctx context.Context, companyID, id uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM shift_templates WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Service) AssignShift(ctx context.Context, employeeID, shiftID uuid.UUID, effectiveDate string, endDate *string) (*EmployeeShift, error) {
	var es EmployeeShift
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO employee_shifts (employee_id, shift_id, effective_date, end_date)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, employee_id, shift_id, effective_date::text, end_date::text, created_at`,
		employeeID, shiftID, effectiveDate, endDate,
	).Scan(&es.ID, &es.EmployeeID, &es.ShiftID, &es.EffectiveDate, &es.EndDate, &es.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &es, nil
}

func (s *Service) ListEmployeeShifts(ctx context.Context, companyID, employeeID uuid.UUID) ([]EmployeeShift, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT es.id, es.employee_id, es.shift_id, es.effective_date::text,
		        es.end_date::text, st.name, es.created_at
		 FROM employee_shifts es
		 JOIN shift_templates st ON st.id = es.shift_id
		 JOIN employees e ON e.id = es.employee_id
		 WHERE e.company_id = $1 AND es.employee_id = $2
		 ORDER BY es.effective_date DESC`, companyID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []EmployeeShift
	for rows.Next() {
		var es EmployeeShift
		if err := rows.Scan(&es.ID, &es.EmployeeID, &es.ShiftID, &es.EffectiveDate,
			&es.EndDate, &es.ShiftName, &es.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, es)
	}
	return list, rows.Err()
}
