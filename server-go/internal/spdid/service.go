package spdid

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

type DIDRecord struct {
	ID         uuid.UUID `json:"id"`
	CompanyID  uuid.UUID `json:"companyId"`
	EmployeeID uuid.UUID `json:"employeeId"`
	DIDString  string    `json:"didString"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

func GenerateDID(companyID, employeeID uuid.UUID) string {
	h := sha256.Sum256([]byte(companyID.String() + employeeID.String()))
	return "did:skillpass:local:" + hex.EncodeToString(h[:])[:16]
}

func (s *Service) CreateDID(ctx context.Context, companyID, employeeID uuid.UUID) (*DIDRecord, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1 AND company_id = $2)`,
		employeeID, companyID,
	).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("employee not found in this company")
	}

	didStr := GenerateDID(companyID, employeeID)

	var r DIDRecord
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO sp_did_records (company_id, employee_id, did_string)
		 VALUES ($1, $2, $3)
		 RETURNING id, company_id, employee_id, did_string, status, created_at`,
		companyID, employeeID, didStr,
	).Scan(&r.ID, &r.CompanyID, &r.EmployeeID, &r.DIDString, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) GetDID(ctx context.Context, companyID, employeeID uuid.UUID) (*DIDRecord, error) {
	var r DIDRecord
	err := s.db.QueryRowContext(ctx,
		`SELECT id, company_id, employee_id, did_string, status, created_at
		 FROM sp_did_records WHERE employee_id = $1 AND company_id = $2`,
		employeeID, companyID,
	).Scan(&r.ID, &r.CompanyID, &r.EmployeeID, &r.DIDString, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Service) RevokeDID(ctx context.Context, companyID, employeeID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE sp_did_records SET status = 'revoked' WHERE employee_id = $1 AND company_id = $2`,
		employeeID, companyID,
	)
	return err
}
