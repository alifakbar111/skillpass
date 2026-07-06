package activity

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type ActivityLog struct {
	ID         uuid.UUID       `json:"id"`
	CompanyID  uuid.UUID       `json:"companyId"`
	ActorID    uuid.UUID       `json:"actorId"`
	ActorName  string          `json:"actorName,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entityType"`
	EntityID   *uuid.UUID      `json:"entityId,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
}

func (s *Service) Log(ctx context.Context, companyID, actorID uuid.UUID, action, entityType string, entityID *uuid.UUID, metadata map[string]any) error {
	metaBytes := []byte("{}")
	if metadata != nil {
		var err error
		metaBytes, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO activity_logs (company_id, actor_id, action, entity_type, entity_id, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6::jsonb)`,
		companyID, actorID, action, entityType, entityID, string(metaBytes),
	)
	return err
}

func (s *Service) List(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]ActivityLog, int, error) {
	var total int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM activity_logs WHERE company_id = $1`, companyID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT al.id, al.company_id, al.actor_id,
		        COALESCE(e.first_name || ' ' || e.last_name, '') as actor_name,
		        al.action, al.entity_type, al.entity_id, al.metadata, al.created_at
		 FROM activity_logs al
		 LEFT JOIN employees e ON e.id = al.actor_id
		 WHERE al.company_id = $1
		 ORDER BY al.created_at DESC
		 LIMIT $2 OFFSET $3`,
		companyID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []ActivityLog
	for rows.Next() {
		var l ActivityLog
		var metaStr string
		if err := rows.Scan(&l.ID, &l.CompanyID, &l.ActorID, &l.ActorName,
			&l.Action, &l.EntityType, &l.EntityID, &metaStr, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		l.Metadata = json.RawMessage(metaStr)
		logs = append(logs, l)
	}
	return logs, total, rows.Err()
}
