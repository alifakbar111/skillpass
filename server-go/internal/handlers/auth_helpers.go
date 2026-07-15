package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/models"
)

func revokeAllForUser(ctx context.Context, tx bun.Tx, userID uuid.UUID) {
	_, _ = tx.NewUpdate().
		Model((*models.RefreshToken)(nil)).
		Set("revoked_at = ?", time.Now()).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Exec(ctx)
}
