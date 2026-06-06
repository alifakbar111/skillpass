package handlers

import (
	"context"
	"database/sql"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/internal/gen"
)

func revokeAllForUser(ctx context.Context, tx *sql.Tx, userID uuid.UUID) {
	stmt := gen.RefreshTokens.UPDATE().SET(
		gen.RefreshTokens.RevokedAt.SET(TimestampzT(time.Now())),
	).WHERE(
		gen.RefreshTokens.UserID.EQ(UUID(userID)).AND(
			gen.RefreshTokens.RevokedAt.IS_NULL(),
		),
	)
	_, _ = stmt.ExecContext(ctx, tx)
}
