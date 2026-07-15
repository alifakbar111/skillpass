package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type RefreshToken struct {
	bun.BaseModel `bun:"table:refresh_tokens"`
	ID        uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	UserID    uuid.UUID  `bun:",notnull"`
	TokenHash string     `bun:",notnull,unique"`
	ExpiresAt time.Time  `bun:",notnull"`
	RevokedAt *time.Time `bun:",nullzero"`
	ReplacedBy *uuid.UUID `bun:",nullzero"`
	CreatedAt time.Time  `bun:",notnull"`
}
