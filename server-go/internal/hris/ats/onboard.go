package ats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// onboardHiredCandidate is the ATS→HRIS account bridge. When a candidate
// accepts an offer we create their employee record and — crucially — LINK it to
// the jobseeker's EXISTING login (matched by email) so they keep one identity
// across the marketplace and the HRIS. Returns whether a user account was linked.
//
// Idempotent: if an employee already exists for this company + email, it is
// reused (and back-linked to the user if not already).
func (s *Service) onboardHiredCandidate(ctx context.Context, candidateID uuid.UUID) (linked bool, err error) {
	var companyID uuid.UUID
	var name, emailAddr string
	var jobPositionTitle sql.NullString
	if err = s.db.QueryRowContext(ctx, `
		SELECT c.company_id, c.candidate_name, c.candidate_email, jp.title
		FROM ats_candidates c
		LEFT JOIN job_postings jp ON jp.id = c.job_posting_id
		WHERE c.id=$1`, candidateID).Scan(&companyID, &name, &emailAddr, &jobPositionTitle); err != nil {
		return false, err
	}
	emailAddr = strings.ToLower(strings.TrimSpace(emailAddr))

	// Match the candidate's existing marketplace login by email.
	var userID uuid.NullUUID
	_ = s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE lower(email)=$1 LIMIT 1`, emailAddr).Scan(&userID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Already an employee for this company + email? Reuse and back-link.
	var existingID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM employees WHERE company_id=$1 AND lower(email)=$2 LIMIT 1`, companyID, emailAddr).Scan(&existingID)
	if err == nil {
		if userID.Valid {
			if _, err = tx.ExecContext(ctx,
				`UPDATE employees SET user_id=$2, updated_at=now() WHERE id=$1 AND user_id IS NULL`, existingID, userID.UUID); err != nil {
				return false, err
			}
		}
		if err = s.assignEmployeeRoleTx(ctx, tx, companyID, existingID); err != nil {
			return false, err
		}
		if err = tx.Commit(); err != nil {
			return false, err
		}
		return userID.Valid, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	// Ensure the company has an employee-id config, then generate the number.
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO employee_id_configs (company_id) VALUES ($1) ON CONFLICT (company_id) DO NOTHING`, companyID); err != nil {
		return false, err
	}
	var prefix string
	var nextVal, padding int
	if err = tx.QueryRowContext(ctx,
		`UPDATE employee_id_configs SET next_sequence = next_sequence + 1 WHERE company_id=$1
		 RETURNING prefix, next_sequence, padding`, companyID).Scan(&prefix, &nextVal, &padding); err != nil {
		return false, err
	}
	empNumber := fmt.Sprintf("%s%0*d", prefix, padding, nextVal)

	first, last := splitName(name)
	var userIDArg any
	if userID.Valid {
		userIDArg = userID.UUID
	}

	var employeeID uuid.UUID
	if err = tx.QueryRowContext(ctx, `
		INSERT INTO employees (company_id, user_id, employee_id_number, first_name, last_name, email, employment_type, employment_status, join_date)
		VALUES ($1,$2,$3,$4,$5,$6,'permanent','active',$7)
		RETURNING id`,
		companyID, userIDArg, empNumber, first, last, emailAddr, time.Now()).Scan(&employeeID); err != nil {
		return false, err
	}

	if err = s.assignEmployeeRoleTx(ctx, tx, companyID, employeeID); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return userID.Valid, nil
}

// assignEmployeeRoleTx grants the company's default "Employee" system role.
func (s *Service) assignEmployeeRoleTx(ctx context.Context, tx *sql.Tx, companyID, employeeID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO employee_roles (employee_id, role_id)
		SELECT $2, r.id FROM hris_roles r
		WHERE r.company_id=$1 AND r.name='Employee' AND r.is_system=true
		ON CONFLICT DO NOTHING`, companyID, employeeID)
	return err
}

func splitName(full string) (first, last string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "New", "Hire"
	}
	parts := strings.SplitN(full, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
