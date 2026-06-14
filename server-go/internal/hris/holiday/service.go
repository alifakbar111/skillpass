package holiday

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

type Holiday struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"companyId"`
	Name        string    `json:"name"`
	Date        string    `json:"date"`
	IsRecurring bool      `json:"isRecurring"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (s *Service) List(ctx context.Context, companyID uuid.UUID, year int) ([]Holiday, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, company_id, name, date::text, is_recurring, created_at
		 FROM holidays WHERE company_id = $1 AND EXTRACT(YEAR FROM date) = $2
		 ORDER BY date`, companyID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Holiday
	for rows.Next() {
		var h Holiday
		if err := rows.Scan(&h.ID, &h.CompanyID, &h.Name, &h.Date, &h.IsRecurring, &h.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, h)
	}
	return list, rows.Err()
}

func (s *Service) Create(ctx context.Context, companyID uuid.UUID, h *Holiday) error {
	return s.db.QueryRowContext(ctx,
		`INSERT INTO holidays (company_id, name, date, is_recurring)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		companyID, h.Name, h.Date, h.IsRecurring,
	).Scan(&h.ID, &h.CreatedAt)
}

func (s *Service) Update(ctx context.Context, companyID, id uuid.UUID, h *Holiday) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE holidays SET name=$1, date=$2, is_recurring=$3
		 WHERE id=$4 AND company_id=$5`,
		h.Name, h.Date, h.IsRecurring, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, companyID, id uuid.UUID) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM holidays WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
