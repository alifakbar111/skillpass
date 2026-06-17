package onboarding

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type Template struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	IsActive    bool           `json:"isActive"`
	Tasks       []TemplateTask `json:"tasks"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type TemplateTask struct {
	ID           uuid.UUID `json:"id"`
	TemplateID   uuid.UUID `json:"templateId"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	SortOrder    int       `json:"sortOrder"`
	DueDays      int       `json:"dueDays"`
	AssigneeRole string    `json:"assigneeRole"`
}

type Checklist struct {
	ID           uuid.UUID       `json:"id"`
	EmployeeID   uuid.UUID       `json:"employeeId"`
	EmployeeName string          `json:"employeeName"`
	EmployeeCode string          `json:"employeeCode"`
	TemplateName string          `json:"templateName"`
	Status       string          `json:"status"`
	StartedAt    time.Time       `json:"startedAt"`
	CompletedAt  *time.Time      `json:"completedAt"`
	Items        []ChecklistItem `json:"items"`
	Progress     int             `json:"progress"`
}

type ChecklistItem struct {
	ID          uuid.UUID  `json:"id"`
	ChecklistID uuid.UUID  `json:"checklistId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	SortOrder   int        `json:"sortOrder"`
	DueDate     *string    `json:"dueDate"`
	IsCompleted bool       `json:"isCompleted"`
	CompletedAt *time.Time `json:"completedAt"`
	CompletedBy *uuid.UUID `json:"completedBy"`
	Notes       string     `json:"notes"`
}

func (s *Service) ListTemplates(ctx context.Context, companyID uuid.UUID) ([]Template, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, is_active, created_at
		FROM onboarding_templates WHERE company_id = $1 ORDER BY name
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Template
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.IsActive, &t.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (s *Service) GetTemplate(ctx context.Context, companyID, templateID uuid.UUID) (*Template, error) {
	t := &Template{}
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, description, is_active, created_at
		FROM onboarding_templates WHERE id = $1 AND company_id = $2
	`, templateID, companyID).Scan(&t.ID, &t.Name, &t.Description, &t.IsActive, &t.CreatedAt)
	if err != nil {
		return nil, err
	}

	taskRows, err := s.db.QueryContext(ctx, `
		SELECT id, template_id, title, description, sort_order, due_days, assignee_role
		FROM onboarding_template_tasks WHERE template_id = $1 ORDER BY sort_order
	`, templateID)
	if err != nil {
		return nil, err
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var task TemplateTask
		if err := taskRows.Scan(&task.ID, &task.TemplateID, &task.Title, &task.Description, &task.SortOrder, &task.DueDays, &task.AssigneeRole); err != nil {
			return nil, err
		}
		t.Tasks = append(t.Tasks, task)
	}
	return t, taskRows.Err()
}

func (s *Service) CreateTemplate(ctx context.Context, companyID uuid.UUID, name, description string, tasks []TemplateTask) (*Template, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	t := &Template{}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO onboarding_templates (company_id, name, description)
		VALUES ($1, $2, $3) RETURNING id, name, description, is_active, created_at
	`, companyID, name, description).Scan(&t.ID, &t.Name, &t.Description, &t.IsActive, &t.CreatedAt)
	if err != nil {
		return nil, err
	}

	for i, task := range tasks {
		var saved TemplateTask
		err = tx.QueryRowContext(ctx, `
			INSERT INTO onboarding_template_tasks (template_id, title, description, sort_order, due_days, assignee_role)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, template_id, title, description, sort_order, due_days, assignee_role
		`, t.ID, task.Title, task.Description, i, task.DueDays, task.AssigneeRole).Scan(
			&saved.ID, &saved.TemplateID, &saved.Title, &saved.Description, &saved.SortOrder, &saved.DueDays, &saved.AssigneeRole)
		if err != nil {
			return nil, err
		}
		t.Tasks = append(t.Tasks, saved)
	}

	return t, tx.Commit()
}

func (s *Service) UpdateTemplate(ctx context.Context, companyID, templateID uuid.UUID, name, description string, isActive bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE onboarding_templates SET name = $3, description = $4, is_active = $5, updated_at = now()
		WHERE id = $1 AND company_id = $2
	`, templateID, companyID, name, description, isActive)
	return err
}

func (s *Service) DeleteTemplate(ctx context.Context, companyID, templateID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM onboarding_templates WHERE id = $1 AND company_id = $2`, templateID, companyID)
	return err
}

func (s *Service) AssignChecklist(ctx context.Context, companyID, employeeID, templateID uuid.UUID) (*Checklist, error) {
	tmpl, err := s.GetTemplate(ctx, companyID, templateID)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var checklistID uuid.UUID
	var startedAt time.Time
	err = tx.QueryRowContext(ctx, `
		INSERT INTO onboarding_checklists (company_id, employee_id, template_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (company_id, employee_id) DO UPDATE SET
			template_id = EXCLUDED.template_id, status = 'in_progress', completed_at = NULL, started_at = now()
		RETURNING id, started_at
	`, companyID, employeeID, templateID).Scan(&checklistID, &startedAt)
	if err != nil {
		return nil, err
	}

	// Clear old items on reassign
	tx.ExecContext(ctx, `DELETE FROM onboarding_checklist_items WHERE checklist_id = $1`, checklistID)

	for i, task := range tmpl.Tasks {
		var dueDate *string
		if task.DueDays > 0 {
			d := startedAt.AddDate(0, 0, task.DueDays).Format("2006-01-02")
			dueDate = &d
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO onboarding_checklist_items (checklist_id, template_task_id, title, description, sort_order, due_date)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, checklistID, task.ID, task.Title, task.Description, i, dueDate)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetChecklist(ctx, companyID, checklistID)
}

func (s *Service) ListChecklists(ctx context.Context, companyID uuid.UUID) ([]Checklist, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT cl.id, cl.employee_id, COALESCE(e.first_name||' '||e.last_name, '') as employee_name, e.employee_id_number,
			COALESCE(t.name, ''), cl.status, cl.started_at, cl.completed_at,
			(SELECT COUNT(*) FILTER (WHERE is_completed) FROM onboarding_checklist_items WHERE checklist_id = cl.id),
			(SELECT COUNT(*) FROM onboarding_checklist_items WHERE checklist_id = cl.id)
		FROM onboarding_checklists cl
		JOIN employees e ON e.id = cl.employee_id
		LEFT JOIN onboarding_templates t ON t.id = cl.template_id
		WHERE cl.company_id = $1
		ORDER BY cl.started_at DESC
	`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Checklist
	for rows.Next() {
		var cl Checklist
		var done, total int
		if err := rows.Scan(&cl.ID, &cl.EmployeeID, &cl.EmployeeName, &cl.EmployeeCode,
			&cl.TemplateName, &cl.Status, &cl.StartedAt, &cl.CompletedAt, &done, &total); err != nil {
			return nil, err
		}
		if total > 0 {
			cl.Progress = done * 100 / total
		}
		result = append(result, cl)
	}
	return result, rows.Err()
}

func (s *Service) GetChecklist(ctx context.Context, companyID, checklistID uuid.UUID) (*Checklist, error) {
	cl := &Checklist{}
	err := s.db.QueryRowContext(ctx, `
		SELECT cl.id, cl.employee_id, COALESCE(e.first_name||' '||e.last_name, '') as employee_name, e.employee_id_number,
			COALESCE(t.name, ''), cl.status, cl.started_at, cl.completed_at
		FROM onboarding_checklists cl
		JOIN employees e ON e.id = cl.employee_id
		LEFT JOIN onboarding_templates t ON t.id = cl.template_id
		WHERE cl.id = $1 AND cl.company_id = $2
	`, checklistID, companyID).Scan(&cl.ID, &cl.EmployeeID, &cl.EmployeeName, &cl.EmployeeCode,
		&cl.TemplateName, &cl.Status, &cl.StartedAt, &cl.CompletedAt)
	if err != nil {
		return nil, err
	}

	itemRows, err := s.db.QueryContext(ctx, `
		SELECT id, checklist_id, title, description, sort_order,
			due_date::text, is_completed, completed_at, completed_by, notes
		FROM onboarding_checklist_items WHERE checklist_id = $1 ORDER BY sort_order
	`, checklistID)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	var done int
	for itemRows.Next() {
		var item ChecklistItem
		if err := itemRows.Scan(&item.ID, &item.ChecklistID, &item.Title, &item.Description,
			&item.SortOrder, &item.DueDate, &item.IsCompleted, &item.CompletedAt, &item.CompletedBy, &item.Notes); err != nil {
			return nil, err
		}
		cl.Items = append(cl.Items, item)
		if item.IsCompleted {
			done++
		}
	}
	if len(cl.Items) > 0 {
		cl.Progress = done * 100 / len(cl.Items)
	}
	return cl, itemRows.Err()
}

func (s *Service) GetMyChecklist(ctx context.Context, companyID, employeeID uuid.UUID) (*Checklist, error) {
	var checklistID uuid.UUID
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM onboarding_checklists WHERE company_id = $1 AND employee_id = $2
	`, companyID, employeeID).Scan(&checklistID)
	if err != nil {
		return nil, err
	}
	return s.GetChecklist(ctx, companyID, checklistID)
}

func (s *Service) CompleteItem(ctx context.Context, companyID, itemID, completedBy uuid.UUID, notes string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Verify item belongs to company
	var checklistID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT ci.checklist_id FROM onboarding_checklist_items ci
		JOIN onboarding_checklists cl ON cl.id = ci.checklist_id
		WHERE ci.id = $1 AND cl.company_id = $2
	`, itemID, companyID).Scan(&checklistID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE onboarding_checklist_items
		SET is_completed = true, completed_at = now(), completed_by = $2, notes = $3
		WHERE id = $1
	`, itemID, completedBy, notes)
	if err != nil {
		return err
	}

	// Auto-complete checklist if all items done
	var remaining int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM onboarding_checklist_items WHERE checklist_id = $1 AND NOT is_completed`,
		checklistID).Scan(&remaining)
	if err != nil {
		return err
	}

	if remaining == 0 {
		_, err = tx.ExecContext(ctx,
			`UPDATE onboarding_checklists SET status = 'completed', completed_at = now() WHERE id = $1`,
			checklistID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) UncompleteItem(ctx context.Context, companyID, completedBy, itemID uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var checklistID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT ci.checklist_id FROM onboarding_checklist_items ci
		JOIN onboarding_checklists cl ON cl.id = ci.checklist_id
		WHERE ci.id = $1 AND cl.company_id = $2
	`, itemID, companyID).Scan(&checklistID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE onboarding_checklist_items
		SET is_completed = false, completed_at = NULL, completed_by = NULL
		WHERE id = $1
	`, itemID)
	if err != nil {
		return err
	}

	// Reopen checklist if it was completed
	_, err = tx.ExecContext(ctx,
		`UPDATE onboarding_checklists SET status = 'in_progress', completed_at = NULL WHERE id = $1 AND status = 'completed'`,
		checklistID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
