// Package authtoken issues and consumes single-use tokens for email
// verification and password reset. Raw tokens go into email links; only
// sha256 hashes are persisted (raw SQL — no go-jet codegen needed).
package authtoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	verificationTTL  = 24 * time.Hour
	passwordResetTTL = 1 * time.Hour
)

// ErrInvalidToken covers unknown, expired, and already-used tokens. Handlers
// must not reveal which case occurred.
var ErrInvalidToken = errors.New("invalid or expired token")

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func newRawToken() (raw, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	raw = hex.EncodeToString(buf)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// CreateEmailVerification issues a verification token for the user and
// returns the raw token to embed in the email link.
func (s *Service) CreateEmailVerification(ctx context.Context, userID string) (string, error) {
	raw, hash, err := newRawToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, hash, time.Now().Add(verificationTTL),
	)
	if err != nil {
		return "", fmt.Errorf("insert verification token: %w", err)
	}
	return raw, nil
}

// ConsumeEmailVerification validates the raw token, marks it used, and flips
// users.is_verified. Returns the user id, or ErrInvalidToken.
func (s *Service) ConsumeEmailVerification(ctx context.Context, rawToken string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var userID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`UPDATE email_verification_tokens
		 SET used_at = now()
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		 RETURNING user_id`,
		hashToken(rawToken),
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidToken
		}
		return "", fmt.Errorf("consume verification token: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET is_verified = TRUE WHERE id = $1`, userID,
	); err != nil {
		return "", fmt.Errorf("mark user verified: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	return userID.String(), nil
}

// PasswordResetUser is the user info needed to address the reset email.
type PasswordResetUser struct {
	UserID string
	Email  string
	Name   string
}

// CreatePasswordReset looks the user up by email and issues a reset token.
// Returns sql.ErrNoRows when no account exists — callers must still respond
// with a generic success message to avoid account enumeration.
func (s *Service) CreatePasswordReset(ctx context.Context, emailAddr string) (string, *PasswordResetUser, error) {
	var u PasswordResetUser
	var userID uuid.UUID
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name FROM users WHERE email = $1`, emailAddr,
	).Scan(&userID, &u.Email, &u.Name)
	if err != nil {
		return "", nil, err // sql.ErrNoRows passes through
	}
	u.UserID = userID.String()

	raw, hash, err := newRawToken()
	if err != nil {
		return "", nil, err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		u.UserID, hash, time.Now().Add(passwordResetTTL),
	)
	if err != nil {
		return "", nil, fmt.Errorf("insert reset token: %w", err)
	}
	return raw, &u, nil
}

// ConsumePasswordReset validates the raw token, marks it used, sets the new
// password hash, and revokes all refresh tokens so stolen sessions die.
func (s *Service) ConsumePasswordReset(ctx context.Context, rawToken, newPasswordHash string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var userID uuid.UUID
	err = tx.QueryRowContext(ctx,
		`UPDATE password_reset_tokens
		 SET used_at = now()
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()
		 RETURNING user_id`,
		hashToken(rawToken),
	).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidToken
		}
		return "", fmt.Errorf("consume reset token: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`, newPasswordHash, userID,
	); err != nil {
		return "", fmt.Errorf("update password: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE refresh_tokens SET revoked_at = now()
		 WHERE user_id = $1 AND revoked_at IS NULL`, userID,
	); err != nil {
		return "", fmt.Errorf("revoke sessions: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}
	return userID.String(), nil
}
