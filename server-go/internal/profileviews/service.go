package profileviews

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ProfileView struct {
	ID              uuid.UUID  `json:"id"`
	ProfileID       uuid.UUID  `json:"profileId"`
	ViewerID        uuid.UUID  `json:"viewerId"`
	CompanyID       *uuid.UUID `json:"companyId,omitempty"`
	ViewedAt        time.Time  `json:"viewedAt"`
	ViewerFirstName string     `json:"viewerFirstName"`
	ViewerLastName  string     `json:"viewerLastName"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) RecordView(ctx context.Context, profileID, viewerID uuid.UUID, companyID *uuid.UUID) error {
	// Prevent duplicate views per day per company
	if companyID != nil {
		var todayExists bool
		_ = s.db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM profile_views WHERE profile_id = $1 AND company_id = $2 AND viewed_at::date = CURRENT_DATE)`,
			profileID, *companyID,
		).Scan(&todayExists)
		if todayExists {
			return nil // already recorded today
		}
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO profile_views (id, profile_id, viewer_id, company_id, viewed_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		uuid.New(), profileID, viewerID, companyID,
	)
	return err
}

func (s *Service) GetViewsByProfile(ctx context.Context, profileID uuid.UUID) ([]ProfileView, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT pv.id, pv.profile_id, pv.viewer_id, pv.company_id, pv.viewed_at,
		        u.first_name, u.last_name
		 FROM profile_views pv
		 JOIN users u ON u.id = pv.viewer_id
		 WHERE pv.profile_id = $1
		 ORDER BY pv.viewed_at DESC`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []ProfileView
	for rows.Next() {
		var v ProfileView
		if err := rows.Scan(&v.ID, &v.ProfileID, &v.ViewerID, &v.CompanyID, &v.ViewedAt, &v.ViewerFirstName, &v.ViewerLastName); err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	return views, rows.Err()
}
